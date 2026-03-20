package usecase

import (
	"errors"
	"testing"

	"anti-bruteforce/internal/domain"
)

type fakeLimiter struct {
	allowed bool
	err     error
}

func (f *fakeLimiter) Allow(ip string, limit int, windowSec int64) (bool, error) {
	return f.allowed, f.err
}

var _ domain.RateLimiter = (*fakeLimiter)(nil)

func TestLoginCheckerAllowed(t *testing.T) {
	limiter := &fakeLimiter{allowed: true}
	checker := NewLoginChecker(limiter, 10, 60)

	res := checker.Check("1.2.3.4")
	if !res.Allowed || res.Exceeded || res.Err != nil {
		t.Fatalf("expected Allowed=true, Exceeded=false, Err=nil, got %+v", res)
	}
}

func TestLoginCheckerExceeded(t *testing.T) {
	limiter := &fakeLimiter{allowed: false}
	checker := NewLoginChecker(limiter, 10, 60)

	res := checker.Check("1.2.3.4")
	if res.Allowed || !res.Exceeded || res.Err != nil {
		t.Fatalf("expected Allowed=false, Exceeded=true, Err=nil, got %+v", res)
	}
}

func TestLoginCheckerError(t *testing.T) {
	limiter := &fakeLimiter{allowed: false, err: errors.New("store failure")}
	checker := NewLoginChecker(limiter, 10, 60)

	res := checker.Check("1.2.3.4")
	if res.Err == nil || res.Allowed || res.Exceeded {
		t.Fatalf("expected Err!=nil, Allowed=false, Exceeded=false, got %+v", res)
	}
}

