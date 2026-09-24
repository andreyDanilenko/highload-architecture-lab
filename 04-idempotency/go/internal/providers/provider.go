package providers

import (
	"context"
	"time"

	"idempotency/internal/domain"
)

type IdempotencyProvider interface {
	// GetRecord returns existing record if exists
	GetRecord(ctx context.Context, key string) (*domain.IdempotencyRecord, error)

	// TryLock attempts to create a new pending record
	// Returns:
	//   - record, nil: lock acquired successfully
	//   - existingRecord, ErrKeyAlreadyProcessing: key already exists and still locked
	//   - nil, error: technical error (redis down, etc)
	TryLock(ctx context.Context, key string, ttl time.Duration) (*domain.IdempotencyRecord, error)

	// Complete stores final result (success or failure)
	Complete(ctx context.Context, key string, result *domain.PaymentResult) error

	// Cleanup removes expired records (optional, can be done by TTL)
	Cleanup(ctx context.Context) error

	// Name returns provider name for logging/metrics
	Name() string
}
