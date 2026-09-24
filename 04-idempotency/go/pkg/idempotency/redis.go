package idempotency

import (
	"context"
	"time"
)

// RedisProvider — subtask 3: SET NX + TTL, общий для всех инстансов.
type RedisProvider struct{}

func NewRedisProvider(_ string, _ time.Duration) (*RedisProvider, error) {
	return nil, ErrProviderNotAvailable
}

func (p *RedisProvider) Name() string { return "redis" }

func (p *RedisProvider) TryLock(_ context.Context, _ string, _ time.Duration) (*Record, error) {
	return nil, ErrProviderNotAvailable
}

func (p *RedisProvider) Complete(_ context.Context, _ string, _ any, _ error) error {
	return ErrProviderNotAvailable
}

func (p *RedisProvider) GetRecord(_ context.Context, _ string) (*Record, error) {
	return nil, ErrRecordNotFound
}

var _ Provider = (*RedisProvider)(nil)
