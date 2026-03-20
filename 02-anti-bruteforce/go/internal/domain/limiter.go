package domain

// RateLimiter is the outbound port: check/record attempts per IP in a sliding window.
// Adapters: in-memory, Redis pessimistic/optimistic/atomic.
type RateLimiter interface {
	// Allow checks the limit and records the attempt if allowed.
	// allowed == false -> 429; err != nil -> 500 (storage error).
	Allow(ip string, limit int, windowSec int64) (allowed bool, err error)
}
