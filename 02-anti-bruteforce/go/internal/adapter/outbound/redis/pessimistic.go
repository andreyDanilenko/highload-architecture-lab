package redisadapter

import (
	"context"
	"fmt"
	"time"

	"anti-bruteforce/internal/domain"
	"github.com/redis/go-redis/v9"
)

// PessimisticLimiter is a Redis-based RateLimiter that uses a per-IP lock.
// It demonstrates the cost of pessimistic distributed locking for rate limiting.
type PessimisticLimiter struct {
	client *redis.Client
}

// NewPessimisticLimiter creates a new Redis pessimistic rate limiter.
func NewPessimisticLimiter(client *redis.Client) *PessimisticLimiter {
	return &PessimisticLimiter{client: client}
}

// Allow implements domain.RateLimiter using a Redis lock per IP.
//
// Algorithm (per docs/subtask-2-pessimistic.md):
//   - Lock key: lock:rate:{ip}
//   - Try to acquire lock via SET NX EX
//   - On success: read/update sorted set rate:{ip}, then release lock
//   - On failure to acquire after retries: treat as storage error.
func (p *PessimisticLimiter) Allow(ip string, limit int, windowSec int64) (bool, error) {
	ctx := context.Background()

	lockKey := fmt.Sprintf("lock:rate:%s", ip)
	// unique-ish lock value; we do not rely on it for ownership checks in this demo
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())

	const (
		lockTTL        = time.Second
		maxLockRetries = 3
		lockRetryDelay = 50 * time.Millisecond
	)

	// Acquire per-IP lock with limited retries.
	acquired := false
	for i := 0; i < maxLockRetries; i++ {
		res, err := p.client.SetArgs(ctx, lockKey, lockValue, redis.SetArgs{
			Mode: "NX",
			TTL:  lockTTL,
		}).Result()
		if err != nil && err != redis.Nil {
			return false, err
		}
		// SET with NX returns "OK" on success, empty string on failure (no error).
		if res == "OK" {
			acquired = true
			break
		}
		time.Sleep(lockRetryDelay)
	}

	if !acquired {
		return false, fmt.Errorf("pessimistic_limiter: failed to acquire lock for ip %s", ip)
	}

	defer func() {
		// Best-effort lock release; ignore errors in this demo.
		_, _ = p.client.Del(ctx, lockKey).Result()
	}()

	now := time.Now()
	cutoff := now.Add(-time.Duration(windowSec) * time.Second).Unix()

	key := fmt.Sprintf("rate:%s", ip)

	// Remove entries strictly older than the window.
	_, err := p.client.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", cutoff-1)).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	// Count entries within the current window.
	count, err := p.client.ZCount(ctx, key, fmt.Sprintf("%d", cutoff), fmt.Sprintf("%d", now.Unix())).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	if count >= int64(limit) {
		return false, nil
	}

	// Add current attempt as score = current unix timestamp.
	score := float64(now.Unix())
	member := fmt.Sprintf("%d", now.UnixNano())

	_, err = p.client.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Result()
	if err != nil {
		return false, err
	}

	return true, nil
}

// Ensure PessimisticLimiter implements domain.RateLimiter at compile time.
var _ domain.RateLimiter = (*PessimisticLimiter)(nil)

