package memory

import (
	"testing"
	"time"
)

func TestStoreAllowWithinLimit(t *testing.T) {
	store := NewStore()
	ip := "1.2.3.4"

	limit := 3
	window := int64(60)

	for i := 0; i < limit; i++ {
		allowed, err := store.Allow(ip, limit, window)
		if err != nil {
			t.Fatalf("unexpected error on attempt %d: %v", i+1, err)
		}
		if !allowed {
			t.Fatalf("expected attempt %d to be allowed", i+1)
		}
	}

	allowed, err := store.Allow(ip, limit, window)
	if err != nil {
		t.Fatalf("unexpected error on attempt over limit: %v", err)
	}
	if allowed {
		t.Fatalf("expected attempt over limit to be rejected")
	}
}

func TestStoreAllowWindowExpires(t *testing.T) {
	store := NewStore()
	ip := "5.6.7.8"

	limit := 1
	window := int64(1) // 1 second

	allowed, err := store.Allow(ip, limit, window)
	if err != nil {
		t.Fatalf("unexpected error on first attempt: %v", err)
	}
	if !allowed {
		t.Fatalf("expected first attempt to be allowed")
	}

	// Immediately after, should be rejected
	allowed, err = store.Allow(ip, limit, window)
	if err != nil {
		t.Fatalf("unexpected error on second attempt: %v", err)
	}
	if allowed {
		t.Fatalf("expected second attempt to be rejected within window")
	}

	// After window passes, should be allowed again
	time.Sleep(1100 * time.Millisecond)

	allowed, err = store.Allow(ip, limit, window)
	if err != nil {
		t.Fatalf("unexpected error after window: %v", err)
	}
	if !allowed {
		t.Fatalf("expected attempt after window to be allowed again")
	}
}

