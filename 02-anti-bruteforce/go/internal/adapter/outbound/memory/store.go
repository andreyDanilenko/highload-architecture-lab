package memory

import (
	"sync"
	"time"

	"anti-bruteforce/internal/domain"
)

// Store is the in-memory adapter for domain.RateLimiter (sliding window).
type Store struct {
	mu   sync.RWMutex
	data map[string][]time.Time
}

// NewStore creates an in-memory attempt store.
func NewStore() *Store {
	return &Store{data: make(map[string][]time.Time)}
}

// Allow implements domain.RateLimiter.
func (s *Store) Allow(ip string, limit int, windowSec int64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Duration(windowSec) * time.Second)

	times := s.data[ip]
	var kept []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}

	if len(kept) >= limit {
		return false, nil
	}

	kept = append(kept, now)
	s.data[ip] = kept
	return true, nil
}

// Ensure Store implements domain.RateLimiter at compile time.
var _ domain.RateLimiter = (*Store)(nil)
