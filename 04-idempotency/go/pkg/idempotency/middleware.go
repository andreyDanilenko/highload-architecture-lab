package idempotency

import (
	"context"
	"time"
)

// Middleware — оркестратор идемпотентности: TryLock → business fn → Complete.
type Middleware struct {
	provider Provider
	lockTTL  time.Duration
}

func NewMiddleware(provider Provider, lockTTL time.Duration) *Middleware {
	return &Middleware{provider: provider, lockTTL: lockTTL}
}

// Execute оборачивает side-effect в протокол идемпотентности.
//
// Поток:
//  1. TryLock — создать pending или вернуть кэш / конфликт
//  2. Если completed/failed — replay без вызова fn
//  3. Иначе выполнить fn и сохранить результат через Complete
func Execute[T any](
	m *Middleware,
	ctx context.Context,
	key string,
	fn func() (*T, error),
) (*T, error) {
	record, err := m.provider.TryLock(ctx, key, m.lockTTL)
	if err != nil {
		return nil, err
	}

	if record.Status == StatusCompleted {
		return castResult[T](record.Result)
	}
	if record.Status == StatusFailed {
		return nil, errFromRecord(record.Error)
	}

	result, businessErr := fn()

	if completeErr := m.provider.Complete(ctx, key, result, businessErr); completeErr != nil {
		// Side-effect уже произошёл — отдаём клиенту результат, логируем потерю кэша
		_ = completeErr
	}

	return result, businessErr
}

func castResult[T any](stored any) (*T, error) {
	if stored == nil {
		return nil, ErrProviderNotAvailable
	}
	typed, ok := stored.(*T)
	if !ok {
		return nil, ErrProviderNotAvailable
	}
	return typed, nil
}

func errFromRecord(msg string) error {
	if msg == "" {
		return nil
	}
	return &storedError{message: msg}
}

type storedError struct {
	message string
}

func (e *storedError) Error() string {
	return e.message
}
