package idempotency

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMemoryConcurrentAcquireHasOneOwner(t *testing.T) {
	p := NewMemoryProvider(10)
	var owners, conflicts atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			claim, err := p.Acquire(context.Background(), "same", "body", time.Minute)
			switch {
			case err == nil && claim.Owner != "":
				owners.Add(1)
			case errors.Is(err, ErrInProgress):
				conflicts.Add(1)
			default:
				t.Errorf("claim=%+v err=%v", claim, err)
			}
		}()
	}
	close(start)
	wg.Wait()
	if owners.Load() != 1 || conflicts.Load() != 99 {
		t.Fatalf("owners=%d conflicts=%d", owners.Load(), conflicts.Load())
	}
}

func TestMemoryRetentionOwnershipAndCopies(t *testing.T) {
	p := NewMemoryProvider(2)
	now := time.Unix(1000, 0)
	p.now = func() time.Time { return now }
	ctx := context.Background()
	claim, err := p.Acquire(ctx, "key", "payload", 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Acquire(ctx, "key", "other", time.Second); !errors.Is(err, ErrMismatch) {
		t.Fatalf("mismatch: %v", err)
	}
	response := Response{Status: 201, Header: http.Header{"Content-Type": {"application/json"}}, Body: []byte("original")}
	if err = p.Complete(ctx, "key", "wrong", response, time.Minute); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("owner: %v", err)
	}
	if err = p.Complete(ctx, "key", claim.Owner, response, 10*time.Second); err != nil {
		t.Fatal(err)
	}
	response.Body[0] = 'X'
	response.Header.Set("Content-Type", "mutated")
	replay, err := p.Acquire(ctx, "key", "payload", time.Second)
	if err != nil || string(replay.Response.Body) != "original" || replay.Response.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("stored aliases caller: %+v %v", replay, err)
	}
	replay.Response.Body[0] = 'Y'
	replay.Response.Header.Set("Content-Type", "changed")
	replay, err = p.Acquire(ctx, "key", "payload", time.Second)
	if err != nil || string(replay.Response.Body) != "original" || replay.Response.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("replay aliases store: %+v %v", replay, err)
	}
	now = now.Add(10 * time.Second)
	fresh, err := p.Acquire(ctx, "key", "new-payload", time.Second)
	if err != nil || fresh.Owner == "" || fresh.Owner == claim.Owner {
		t.Fatalf("expired complete did not release key: %+v %v", fresh, err)
	}
	if err = p.Complete(ctx, "key", claim.Owner, response, time.Second); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("stale owner: %v", err)
	}
}

func TestMemoryExpiredPendingQuarantinedAndCapacity(t *testing.T) {
	p := NewMemoryProvider(1)
	now := time.Unix(1000, 0)
	p.now = func() time.Time { return now }
	ctx := context.Background()
	claim, err := p.Acquire(ctx, "key", "body", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Second)
	if _, err = p.Acquire(ctx, "key", "body", time.Second); !errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("expired pending: %v", err)
	}
	if err = p.Complete(ctx, "key", claim.Owner, Response{Status: 200}, time.Minute); !errors.Is(err, ErrOutcomeUnknown) {
		t.Fatalf("late completion: %v", err)
	}
	if _, err = p.Acquire(ctx, "other", "body", time.Second); !errors.Is(err, ErrCapacity) {
		t.Fatalf("quarantine must not be silently evicted: %v", err)
	}
}

func TestMemoryExpiredCompletedReclaimsCapacity(t *testing.T) {
	p := NewMemoryProvider(1)
	now := time.Unix(1000, 0)
	p.now = func() time.Time { return now }
	ctx := context.Background()
	claim, err := p.Acquire(ctx, "a", "body", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Complete(ctx, "a", claim.Owner, Response{Status: 200}, time.Second); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Second)
	if _, err = p.Acquire(ctx, "b", "body", time.Second); err != nil {
		t.Fatalf("reclaim: %v", err)
	}
}

func TestProviderCancellationAndNoopBaseline(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, p := range []Provider{NewMemoryProvider(10), NewNoopProvider()} {
		if _, err := p.Acquire(cancelled, "key", "body", time.Second); !errors.Is(err, context.Canceled) {
			t.Fatalf("%s: %v", p.Name(), err)
		}
	}
	p := NewNoopProvider()
	ctx := context.Background()
	a, err := p.Acquire(ctx, "key", "body", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Complete(ctx, "key", a.Owner, Response{Status: 200}, time.Minute); err != nil {
		t.Fatal(err)
	}
	b, err := p.Acquire(ctx, "key", "body", time.Second)
	if err != nil || b.Response != nil || b.Owner == a.Owner {
		t.Fatalf("noop must repeat: %+v %v", b, err)
	}
}
