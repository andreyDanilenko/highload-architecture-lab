package idempotency

import "errors"

var (
	ErrKeyAlreadyProcessing = errors.New("request with this key is already processing")
	ErrRecordNotFound       = errors.New("idempotency record not found")
	ErrProviderNotAvailable = errors.New("idempotency provider not available")
)
