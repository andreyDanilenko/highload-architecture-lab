package idempotency

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"labshared/httpx"
)

// HTTPOptions задаёт отдельно срок владения, окно повторов и пределы памяти.
type HTTPOptions struct {
	LeaseTTL         time.Duration
	ResultTTL        time.Duration
	CompleteTimeout  time.Duration
	MaxBodyBytes     int64
	MaxResponseBytes int64
	Logger           *slog.Logger
}

// HTTPMiddleware сохраняет законченный HTTP-ответ, а не указатель на доменный объект.
// Применяется к конкретному маршруту записи; streaming и произвольные заголовки не поддерживаются.
type HTTPMiddleware struct {
	provider Provider
	options  HTTPOptions
	outcomes [8]atomic.Uint64
	duration atomic.Uint64
}

var outcomeNames = [...]string{"executed", "replayed", "in_progress", "mismatch", "unknown", "rejected", "unavailable", "panic"}

func NewHTTPMiddleware(provider Provider, options HTTPOptions) (*HTTPMiddleware, error) {
	if provider == nil {
		return nil, errors.New("idempotency provider is required")
	}
	if options.LeaseTTL <= 0 || options.ResultTTL <= 0 || options.CompleteTimeout <= 0 || options.MaxBodyBytes <= 0 || options.MaxResponseBytes <= 0 {
		return nil, errors.New("idempotency durations and body limits must be positive")
	}
	if options.MaxBodyBytes > 64<<20 || options.MaxResponseBytes > 64<<20 {
		return nil, errors.New("body limits must not exceed 64 MiB")
	}
	if options.Logger == nil {
		options.Logger = slog.Default()
	}
	return &HTTPMiddleware{provider: provider, options: options}, nil
}

func (m *HTTPMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		outcome := 5
		defer func() { m.outcomes[outcome].Add(1); m.duration.Add(uint64(time.Since(started))) }()

		// Шаг 1: валидируем контракт до захвата; ошибочный ключ не занимает хранилище.
		keys := r.Header.Values("Idempotency-Key")
		if len(keys) != 1 || !visibleASCII(keys[0], 128) {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_key", "one Idempotency-Key of 1..128 visible ASCII characters is required")
			return
		}
		tenant := "lab"
		if values := r.Header.Values("X-Demo-Tenant"); len(values) != 0 {
			if len(values) != 1 || !visibleASCII(values[0], 64) {
				httpx.WriteError(w, http.StatusBadRequest, "invalid_tenant", "X-Demo-Tenant must contain 1..64 visible ASCII characters")
				return
			}
			tenant = values[0]
		}
		// X-Demo-Tenant показывает разделение ключей в лаборатории, но не аутентифицирует пользователя.
		raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, m.options.MaxBodyBytes))
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				httpx.WriteError(w, http.StatusRequestEntityTooLarge, "body_too_large", "request body exceeds the configured limit")
			} else {
				httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "request body could not be read")
			}
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(raw))
		route := r.Pattern
		if route == "" {
			route = r.URL.EscapedPath()
		}
		key := digest([]byte(r.Method + "\x00" + route + "\x00" + tenant + "\x00" + keys[0]))
		// Pattern объединяет маршрут, но path/query могут менять смысл операции.
		// Сравниваем их и исходные байты тела без нормализации JSON или query.
		fingerprintInput := append([]byte(r.URL.EscapedPath()+"\x00"+r.URL.RawQuery+"\x00"), raw...)
		fingerprint := digest(fingerprintInput)

		// Шаг 2: атомарный захват решает, кто исполняет запрос, а кто получает replay или конфликт.
		claim, err := m.provider.Acquire(r.Context(), key, fingerprint, m.options.LeaseTTL)
		if err != nil {
			outcome = m.writeProviderError(w, err)
			return
		}
		if claim.Response != nil {
			outcome = 1
			writeResponse(w, *claim.Response)
			return
		}

		// Шаг 3: буфер не отправляет клиенту успех до сохранения результата.
		recorder := &responseBuffer{header: make(http.Header), limit: m.options.MaxResponseBytes}
		panicked := true
		func() {
			defer func() {
				if panicked {
					_ = recover()
				}
			}()
			next.ServeHTTP(recorder, r)
			panicked = false
		}()
		if panicked {
			outcome = 7
			m.options.Logger.Error("idempotency handler panic; outcome requires reconciliation", "key_hash", key)
			httpx.WriteError(w, http.StatusInternalServerError, "outcome_unknown", "operation outcome is unknown; do not retry with a new key")
			return // pending сохраняется: повтор не должен повторить возможный эффект.
		}
		response, ok := recorder.response()
		if !ok {
			outcome = 4
			m.options.Logger.Error("idempotency response exceeded limit", "key_hash", key)
			httpx.WriteError(w, http.StatusServiceUnavailable, "outcome_unknown", "operation outcome is unknown; do not retry with a new key")
			return
		}

		// Шаг 4: сохраняем любой законченный ответ handler, включая 400/408/503.
		// Retry с тем же ключом воспроизведёт ошибку; это явная политика этого API.
		// Отмена клиентом не должна мешать сохранить уже известный результат.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), m.options.CompleteTimeout)
		defer cancel()
		if err := m.provider.Complete(ctx, key, claim.Owner, response, m.options.ResultTTL); err != nil {
			outcome = 4
			// Не выводим тела, ключи и тексты инфраструктурных ошибок в ответ или лог.
			m.options.Logger.Error("idempotency completion failed; outcome requires reconciliation", "key_hash", key)
			httpx.WriteError(w, http.StatusServiceUnavailable, "outcome_unknown", "operation outcome is unknown; do not retry with a new key")
			return
		}
		outcome = 0
		writeResponse(w, response)
	})
}

func (m *HTTPMiddleware) writeProviderError(w http.ResponseWriter, err error) int {
	switch {
	case errors.Is(err, ErrInProgress):
		httpx.WriteConflict(w, "in_progress", "operation is still processing", 1)
		return 2
	case errors.Is(err, ErrMismatch):
		httpx.WriteError(w, http.StatusConflict, "key_mismatch", "this key was used with a different request body")
		return 3
	case errors.Is(err, ErrOutcomeUnknown), errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, http.StatusConflict, "outcome_unknown", "operation outcome is unknown; do not retry with a new key")
		return 4
	default:
		httpx.WriteError(w, http.StatusServiceUnavailable, "idempotency_unavailable", "operation could not be admitted; retry later with the same key")
		return 6
	}
}

func visibleASCII(value string, max int) bool {
	if len(value) == 0 || len(value) > max {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < 33 || value[i] > 126 {
			return false
		}
	}
	return true
}
func digest(value []byte) string { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }

// MetricsHandler использует фиксированные labels: ключи/tenant/body никогда не становятся labels.
func (m *HTTPMiddleware) MetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = fmt.Fprintln(w, "# HELP idempotency_requests_total Requests by middleware outcome.\n# TYPE idempotency_requests_total counter")
		var total uint64
		for i, name := range outcomeNames {
			n := m.outcomes[i].Load()
			total += n
			_, _ = fmt.Fprintf(w, "idempotency_requests_total{outcome=%q} %d\n", name, n)
		}
		_, _ = fmt.Fprintf(w, "# HELP idempotency_duration_seconds Middleware duration including operation.\n# TYPE idempotency_duration_seconds summary\nidempotency_duration_seconds_sum %g\nidempotency_duration_seconds_count %d\n", float64(m.duration.Load())/float64(time.Second), total)
	})
}

type responseBuffer struct {
	header       http.Header
	frozenHeader http.Header
	status       int
	body         bytes.Buffer
	limit        int64
	overflow     bool
}

func (b *responseBuffer) Header() http.Header { return b.header }
func (b *responseBuffer) WriteHeader(status int) {
	if status < 100 || status > 999 {
		panic("invalid HTTP status")
	}
	if status < 200 || b.status != 0 {
		return
	}
	b.status = status
	b.frozenHeader = make(http.Header)
	for _, name := range []string{"Content-Type", "Location"} {
		for _, value := range b.header.Values(name) {
			b.frozenHeader.Add(name, value)
		}
	}
}
func (b *responseBuffer) Write(data []byte) (int, error) {
	if b.status == 0 {
		b.WriteHeader(http.StatusOK)
	}
	if b.overflow || int64(b.body.Len())+int64(len(data)) > b.limit {
		b.overflow = true
		return 0, errors.New("idempotency response limit exceeded")
	}
	return b.body.Write(data)
}
func (b *responseBuffer) response() (Response, bool) {
	if b.status == 0 {
		b.WriteHeader(http.StatusOK)
	}
	response := Response{Status: b.status, Header: make(http.Header), Body: bytes.Clone(b.body.Bytes())}
	size := int64(len(response.Body))
	// Сохраняем только разрешённые endpoint-заголовки; cookies и hop-by-hop не воспроизводятся.
	for _, name := range []string{"Content-Type", "Location"} {
		for _, value := range b.frozenHeader.Values(name) {
			size += int64(len(name) + len(value))
			response.Header.Add(name, value)
		}
	}
	return response, !b.overflow && size <= b.limit
}
func writeResponse(w http.ResponseWriter, response Response) {
	for _, name := range []string{"Content-Type", "Location"} {
		for _, value := range response.Header.Values(name) {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(response.Status)
	_, _ = w.Write(response.Body)
}
