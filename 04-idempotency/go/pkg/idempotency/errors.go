package idempotency

import "errors"

var (
	ErrInProgress      = errors.New("operation is in progress")
	ErrMismatch        = errors.New("key was used with different request parameters")
	ErrOutcomeUnknown  = errors.New("operation outcome is unknown")
	ErrCapacity        = errors.New("idempotency store capacity reached")
	ErrNotOwner        = errors.New("attempt no longer owns the record")
	ErrUnavailable     = errors.New("idempotency store unavailable")
	ErrInvalidArgument = errors.New("invalid idempotency argument")
)
