package idempotency_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"idempotency/pkg/idempotency"
)

func httpOptions() idempotency.HTTPOptions {
	return idempotency.HTTPOptions{LeaseTTL: time.Minute, ResultTTL: time.Hour, CompleteTimeout: time.Second, MaxBodyBytes: 1024, MaxResponseBytes: 1024, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
}
func httpMiddleware(t *testing.T, p idempotency.Provider, o idempotency.HTTPOptions) *idempotency.HTTPMiddleware {
	t.Helper()
	m, err := idempotency.NewHTTPMiddleware(p, o)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func httpRequest(key, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/payments", strings.NewReader(body))
	if key != "" {
		r.Header.Set("Idempotency-Key", key)
	}
	return r
}
func httpServe(h http.Handler, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestHTTPConcurrentRequestsAndExactReplay(t *testing.T) {
	m := httpMiddleware(t, idempotency.NewMemoryProvider(100), httpOptions())
	entered, release := make(chan struct{}), make(chan struct{})
	var effects atomic.Int32
	h := m.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		effects.Add(1)
		close(entered)
		<-release
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", "/payments/1")
		w.Header().Set("Set-Cookie", "session=private")
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Location", "/too-late")
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	first := make(chan *httptest.ResponseRecorder, 1)
	go func() { first <- httpServe(h, httpRequest("same", `{"amount":1}`)) }()
	<-entered
	var wg sync.WaitGroup
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got := httpServe(h, httpRequest("same", `{"amount":1}`))
			if got.Code != http.StatusConflict || !strings.Contains(got.Body.String(), "in_progress") {
				t.Errorf("pending: %d %s", got.Code, got.Body.String())
			}
		}()
	}
	wg.Wait()
	close(release)
	original := <-first
	replay := httpServe(h, httpRequest("same", `{"amount":1}`))
	if original.Code != http.StatusCreated || replay.Code != original.Code || replay.Body.String() != original.Body.String() {
		t.Fatalf("replay differs: %d/%d %s/%s", original.Code, replay.Code, original.Body.String(), replay.Body.String())
	}
	if replay.Header().Get("Content-Type") != "application/json" || replay.Header().Get("Location") != "/payments/1" || replay.Header().Get("Set-Cookie") != "" {
		t.Fatalf("unexpected replay headers: %v", replay.Header())
	}
	if effects.Load() != 1 {
		t.Fatalf("effects=%d", effects.Load())
	}
	metrics := httpServe(m.MetricsHandler(), httptest.NewRequest("GET", "/metrics", nil)).Body.String()
	for _, want := range []string{`idempotency_requests_total{outcome="executed"} 1`, `idempotency_requests_total{outcome="replayed"} 1`, `idempotency_requests_total{outcome="in_progress"} 12`} {
		if !strings.Contains(metrics, want) {
			t.Errorf("missing metric %q", want)
		}
	}
}

func TestHTTPMismatchAndScope(t *testing.T) {
	m := httpMiddleware(t, idempotency.NewMemoryProvider(100), httpOptions())
	var calls atomic.Int32
	h := m.Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); _, _ = w.Write([]byte("ok")) }))
	if got := httpServe(h, httpRequest("key", "one")); got.Code != 200 {
		t.Fatal(got.Code)
	}
	if got := httpServe(h, httpRequest("key", "two")); got.Code != 409 || !strings.Contains(got.Body.String(), "key_mismatch") {
		t.Fatalf("mismatch: %d %s", got.Code, got.Body.String())
	}
	tenant := httpRequest("key", "two")
	tenant.Header.Set("X-Demo-Tenant", "other")
	if got := httpServe(h, tenant); got.Code != 200 {
		t.Fatal(got.Code)
	}
	otherRoute := httpRequest("key", "two")
	otherRoute.URL.Path = "/other"
	if got := httpServe(h, otherRoute); got.Code != 200 {
		t.Fatal(got.Code)
	}
	if calls.Load() != 3 {
		t.Fatalf("scoped effects=%d", calls.Load())
	}
}

func TestHTTPRejectsInvalidInputBeforeAcquire(t *testing.T) {
	tests := []struct {
		name, key, body, tenant string
		duplicate               bool
		want                    int
	}{
		{name: "missing", body: "{}", want: 400},
		{name: "space", key: "has space", body: "{}", want: 400},
		{name: "too long", key: strings.Repeat("a", 129), body: "{}", want: 400},
		{name: "non ascii", key: "ключ", body: "{}", want: 400},
		{name: "duplicate", key: "x", body: "{}", duplicate: true, want: 400},
		{name: "tenant", key: "x", body: "{}", tenant: "bad tenant", want: 400},
		{name: "large body", key: "x", body: strings.Repeat("x", 9), want: 413},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &httpFakeProvider{acquire: func(context.Context, string, string, time.Duration) (idempotency.Claim, error) {
				t.Fatal("invalid input acquired a key")
				return idempotency.Claim{}, nil
			}}
			o := httpOptions()
			o.MaxBodyBytes = 8
			m := httpMiddleware(t, p, o)
			h := m.Wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("invalid input executed") }))
			r := httpRequest(tt.key, tt.body)
			if tt.tenant != "" {
				r.Header.Set("X-Demo-Tenant", tt.tenant)
			}
			if tt.duplicate {
				r.Header.Add("Idempotency-Key", "second")
			}
			got := httpServe(h, r)
			if got.Code != tt.want {
				t.Fatalf("status=%d, body=%s", got.Code, got.Body.String())
			}
		})
	}
}

func TestHTTPCompletionUsesDetachedBoundedContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var saved bool
	p := &httpFakeProvider{complete: func(ctx context.Context, _, _ string, r idempotency.Response, _ time.Duration) error {
		if ctx.Err() != nil {
			t.Fatalf("completion inherited cancellation: %v", ctx.Err())
		}
		deadline, ok := ctx.Deadline()
		if !ok || time.Until(deadline) > time.Second {
			t.Fatal("completion lacks bounded deadline")
		}
		if r.Status != 201 || string(r.Body) != "effect" {
			t.Fatalf("unexpected saved response: %+v", r)
		}
		saved = true
		return nil
	}}
	h := httpMiddleware(t, p, httpOptions()).Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cancel()
		w.WriteHeader(201)
		_, _ = w.Write([]byte("effect"))
	}))
	got := httpServe(h, httpRequest("key", "{}").WithContext(ctx))
	if got.Code != 201 || !saved {
		t.Fatalf("result=%d saved=%v", got.Code, saved)
	}
}

func TestHTTPCompletionFailureDoesNotSendSuccess(t *testing.T) {
	p := &httpFakeProvider{complete: func(context.Context, string, string, idempotency.Response, time.Duration) error {
		return errors.New("secret Redis address")
	}}
	h := httpMiddleware(t, p, httpOptions()).Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(201)
		_, _ = w.Write([]byte("payment succeeded"))
	}))
	got := httpServe(h, httpRequest("key", "{}"))
	if got.Code != 503 || !strings.Contains(got.Body.String(), "outcome_unknown") || strings.Contains(got.Body.String(), "secret") || strings.Contains(got.Body.String(), "payment succeeded") {
		t.Fatalf("result: %d %s", got.Code, got.Body.String())
	}
}

func TestHTTPPanicAndOversizedResponseLeavePending(t *testing.T) {
	for _, kind := range []string{"panic", "body", "header"} {
		t.Run(kind, func(t *testing.T) {
			o := httpOptions()
			o.MaxResponseBytes = 64
			var calls int
			h := httpMiddleware(t, idempotency.NewMemoryProvider(10), o).Wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls++
				switch kind {
				case "panic":
					panic("private panic")
				case "body":
					_, _ = w.Write([]byte(strings.Repeat("x", 65)))
				case "header":
					w.Header().Set("Location", strings.Repeat("x", 65))
					w.WriteHeader(201)
				}
			}))
			first := httpServe(h, httpRequest("key", "{}"))
			want := 503
			if kind == "panic" {
				want = 500
			}
			if first.Code != want || !strings.Contains(first.Body.String(), "outcome_unknown") {
				t.Fatalf("first=%d %s", first.Code, first.Body.String())
			}
			retry := httpServe(h, httpRequest("key", "{}"))
			if retry.Code != 409 || calls != 1 {
				t.Fatalf("retry=%d calls=%d", retry.Code, calls)
			}
		})
	}
}

func TestHTTPProviderErrorsDoNotLeakOrExecute(t *testing.T) {
	for _, test := range []struct {
		err    error
		status int
		code   string
	}{
		{idempotency.ErrMismatch, 409, "key_mismatch"},
		{idempotency.ErrOutcomeUnknown, 409, "outcome_unknown"},
		{idempotency.ErrCapacity, 503, "idempotency_unavailable"},
		{idempotency.ErrUnavailable, 503, "idempotency_unavailable"},
		{errors.New("secret"), 503, "idempotency_unavailable"},
	} {
		t.Run(test.code+test.err.Error(), func(t *testing.T) {
			p := &httpFakeProvider{acquire: func(context.Context, string, string, time.Duration) (idempotency.Claim, error) {
				return idempotency.Claim{}, test.err
			}}
			h := httpMiddleware(t, p, httpOptions()).Wrap(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("executed after failed acquire") }))
			got := httpServe(h, httpRequest("key", "{}"))
			if got.Code != test.status || !strings.Contains(got.Body.String(), test.code) || strings.Contains(got.Body.String(), "secret") {
				t.Fatalf("%d %s", got.Code, got.Body.String())
			}
		})
	}
}

type httpFakeProvider struct {
	acquire  func(context.Context, string, string, time.Duration) (idempotency.Claim, error)
	complete func(context.Context, string, string, idempotency.Response, time.Duration) error
}

func (p *httpFakeProvider) Name() string { return "fake" }
func (p *httpFakeProvider) Acquire(ctx context.Context, key, fingerprint string, lease time.Duration) (idempotency.Claim, error) {
	if p.acquire != nil {
		return p.acquire(ctx, key, fingerprint, lease)
	}
	return idempotency.Claim{Owner: "owner"}, nil
}
func (p *httpFakeProvider) Complete(ctx context.Context, key, owner string, response idempotency.Response, retention time.Duration) error {
	if p.complete != nil {
		return p.complete(ctx, key, owner, response, retention)
	}
	return nil
}

// ServeMux устанавливает один Pattern для разных значений wildcard: payload обязан учитывать конкретный ресурс.
func TestHTTPFingerprintIncludesConcretePathAndQuery(t *testing.T) {
	m := httpMiddleware(t, idempotency.NewMemoryProvider(10), httpOptions())
	var effects int
	mux := http.NewServeMux()
	mux.Handle("POST /accounts/{account}/payments", m.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		effects++
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(r.PathValue("account") + ":" + r.URL.RawQuery))
	})))
	send := func(target string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, target, strings.NewReader(`{"amount":1}`))
		r.Header.Set("Idempotency-Key", "same-key")
		return httpServe(mux, r)
	}
	first := send("/accounts/alice/payments?currency=USD")
	for _, target := range []string{"/accounts/bob/payments?currency=USD", "/accounts/alice/payments?currency=EUR"} {
		got := send(target)
		if got.Code != http.StatusConflict || !strings.Contains(got.Body.String(), "key_mismatch") {
			t.Errorf("changed resource %s: %d %s", target, got.Code, got.Body.String())
		}
	}
	replay := send("/accounts/alice/payments?currency=USD")
	if first.Code != 200 || replay.Code != 200 || first.Body.String() != replay.Body.String() || effects != 1 {
		t.Fatalf("replay: statuses=%d/%d effects=%d bodies=%q/%q", first.Code, replay.Code, effects, first.Body.String(), replay.Body.String())
	}
}
