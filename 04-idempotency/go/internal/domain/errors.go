package domain

import "errors"

var (
	ErrIdempotencyKeyRequired = errors.New("idempotency key is required")
	ErrKeyAlreadyProcessing   = errors.New("request with this key is already processing")
	ErrProviderNotAvailable   = errors.New("idempotency provider not available")
	ErrRecordNotFound         = errors.New("idempotency record not found")
)
