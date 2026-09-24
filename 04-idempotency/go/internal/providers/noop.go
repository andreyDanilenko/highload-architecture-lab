package providers

import (
	"context"
	"time"

	"idempotency/internal/domain"
)

// NoopProvider реализует IdempotencyProvider, но НЕ хранит ничего.
// Каждый запрос считается новым — демонстрирует проблему.
type NoopProvider struct{}

func NewNoopProvider() *NoopProvider {
	return &NoopProvider{}
}

func (p *NoopProvider) Name() string {
	return "noop"
}

// GetRecord всегда возвращает ErrRecordNotFound — запись никогда не существует
func (p *NoopProvider) GetRecord(ctx context.Context, key string) (*domain.IdempotencyRecord, error) {
	return nil, domain.ErrRecordNotFound
}

// TryLock всегда создает новую запись (как будто lock всегда успешен)
// Возвращает новую pending-запись
func (p *NoopProvider) TryLock(ctx context.Context, key string, ttl time.Duration) (*domain.IdempotencyRecord, error) {
	// Всегда говорим "да, lock взят, выполняй операцию"
	return domain.NewIdempotencyRecord(key, ttl), nil
}

// Complete ничего не сохраняет (просто заглушка)
func (p *NoopProvider) Complete(ctx context.Context, key string, result *domain.PaymentResult) error {
	// Ничего не делаем — результат не сохраняется
	return nil
}

// Cleanup — ничего не чистим
func (p *NoopProvider) Cleanup(ctx context.Context) error {
	return nil
}

// Проверка интерфейса на этапе компиляции
var _ IdempotencyProvider = (*NoopProvider)(nil)
