package usecase

import "anti-bruteforce/internal/domain"

// CheckResult is the result of a login attempt check (maps to HTTP 200/429/500).
type CheckResult struct {
	Allowed  bool  // -> 200
	Exceeded bool  // -> 429
	Err      error // -> 500
}

// LoginChecker is the use case port: check login attempt by IP.
type LoginChecker interface {
	Check(ip string) CheckResult
}

// loginChecker implements the use case using domain.RateLimiter.
type loginChecker struct {
	limiter domain.RateLimiter
	limit   int
	window  int64
}

// NewLoginChecker creates the use case with the given limiter and window params.
func NewLoginChecker(limiter domain.RateLimiter, limit int, windowSec int64) LoginChecker {
	return &loginChecker{limiter: limiter, limit: limit, window: windowSec}
}

func (u *loginChecker) Check(ip string) CheckResult {
	allowed, err := u.limiter.Allow(ip, u.limit, u.window)
	if err != nil {
		return CheckResult{Err: err}
	}
	if !allowed {
		return CheckResult{Exceeded: true}
	}
	return CheckResult{Allowed: true}
}
