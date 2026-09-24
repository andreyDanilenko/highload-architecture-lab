package idempotency

import (
	"context"
	"time"
)

// NoopProvider — контрольный опыт: каждый повтор снова допускается к эффекту.
type NoopProvider struct{}

func NewNoopProvider() *NoopProvider { return &NoopProvider{} }
func (*NoopProvider) Name() string   { return "noop" }

func (*NoopProvider) Acquire(ctx context.Context, _, _ string, _ time.Duration) (Claim, error) {
	if err := ctx.Err(); err != nil {
		return Claim{}, err
	}
	owner, err := newOwner()
	return Claim{Owner: owner}, err
}

func (*NoopProvider) Complete(ctx context.Context, _, _ string, _ Response, _ time.Duration) error {
	return ctx.Err()
}

var _ Provider = (*NoopProvider)(nil)
