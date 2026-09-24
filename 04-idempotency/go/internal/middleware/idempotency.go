package middleware

import (
	"context"
	"net/http"
	"time"

	"idempotency/internal/domain"
	"idempotency/internal/providers"
)

type IdempotencyMiddleware struct {
	provider providers.IdempotencyProvider
	lockTTL  time.Duration
}

func NewIdempotencyMiddleware(provider providers.IdempotencyProvider, lockTTL time.Duration) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{
		provider: provider,
		lockTTL:  lockTTL,
	}
}

// ExecuteWithIdempotency — core logic
// Если запрос уже был выполнен — возвращает сохраненный результат
// Если нет — выполняет fn и сохраняет результат
func (m *IdempotencyMiddleware) ExecuteWithIdempotency(
	ctx context.Context,
	key string,
	fn func() (*domain.PaymentResult, error),
) (*domain.PaymentResult, error) {
	// Пытаемся захватить lock
	record, err := m.provider.TryLock(ctx, key, m.lockTTL)
	if err != nil {
		// Если ключ уже в обработке
		if err == domain.ErrKeyAlreadyProcessing {
			return nil, err
		}
		// Техническая ошибка
		return nil, err
	}

	// Если запись уже завершена (была в кэше)
	if record.Status == domain.StatusCompleted || record.Status == domain.StatusFailed {
		return record.Result, nil
	}

	// Если record.Status == StatusPending — это первый запрос, выполняем
	result, err := fn()

	// Сохраняем результат
	var saveErr error
	if err != nil {
		saveErr = m.provider.Complete(ctx, key, domain.NewFailedPayment(err.Error()))
	} else {
		saveErr = m.provider.Complete(ctx, key, result)
	}

	if saveErr != nil {
		// Логируем, но возвращаем результат операции
		// TODO: добавить логирование
	}

	return result, err
}

type contextKey string

const idempotencyKeyKey contextKey = "idempotency_key"

// Middleware HTTP для извлечения ключа из заголовка
func (m *IdempotencyMiddleware) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			// Пропускаем, но handler сам вернет 400
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), idempotencyKeyKey, key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetIdempotencyKey(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(idempotencyKeyKey).(string)
	return key, ok
}
