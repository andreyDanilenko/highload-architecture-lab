package idempotency

import (
	"context"
	"time"
)

// NoopProvider — subtask 1: идемпотентности нет, каждый retry = новый side-effect.
type NoopProvider struct{}

func NewNoopProvider() *NoopProvider {
	return &NoopProvider{}
}

func (p *NoopProvider) Name() string { return "noop" }

func (p *NoopProvider) TryLock(_ context.Context, key string, lockTTL time.Duration) (*Record, error) {
	return NewRecord(key, lockTTL), nil
}

func (p *NoopProvider) Complete(_ context.Context, _ string, _ any, _ error) error {
	return nil
}

func (p *NoopProvider) GetRecord(_ context.Context, _ string) (*Record, error) {
	return nil, ErrRecordNotFound
}

var _ Provider = (*NoopProvider)(nil)
