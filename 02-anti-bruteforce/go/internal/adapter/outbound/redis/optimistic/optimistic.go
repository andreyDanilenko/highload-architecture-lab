package optimistic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"anti-bruteforce/internal/domain"

	"github.com/redis/go-redis/v9"
)

// errOverLimit is returned from the Watch callback when limit is exceeded.
// We use a sentinel to distinguish "429" from "retry" and "real error".
var errOverLimit = errors.New("over limit")

// OptimisticLimiter is a Redis-based RateLimiter using WATCH/MULTI/EXEC.
// No locks: read → check → write. If key changed between read and write, retry.
type OptimisticLimiter struct {
	client *redis.Client
}

// NewOptimisticLimiter creates a new Redis optimistic rate limiter.
func NewOptimisticLimiter(client *redis.Client) *OptimisticLimiter {
	return &OptimisticLimiter{client: client}
}

// Allow implements domain.RateLimiter using Redis WATCH + TxPipelined.
//
// Algorithm (per docs/subtask-3-optimistic.md):
//   - WATCH rate:{ip}
//   - ZRANGE to get entries, filter by window client-side, count
//   - If count >= limit: return false (429)
//   - MULTI; ZREMRANGEBYSCORE (trim old); ZADD (new attempt); EXEC
//   - If EXEC fails (key modified): retry up to maxRetries
func (o *OptimisticLimiter) Allow(ip string, limit int, windowSec int64) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("rate:%s", ip)
	now := time.Now()
	cutoff := now.Add(-time.Duration(windowSec) * time.Second).Unix()
	cutoffStr := fmt.Sprintf("%d", cutoff-1)

	score := float64(now.Unix())
	member := fmt.Sprintf("%d", now.UnixNano())

	const maxRetries = 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := o.client.Watch(ctx, func(tx *redis.Tx) error {
			// 1. Read current entries
			entries, err := tx.ZRangeWithScores(ctx, key, 0, -1).Result()
			if err != nil && err != redis.Nil {
				return err
			}

			// 2. Filter by sliding window (client-side)
			var count int
			for _, z := range entries {
				if z.Score >= float64(cutoff) {
					count++
				}
			}

			if count >= limit {
				return errOverLimit
			}

			// 3. Queue: trim old + add new (atomic in one transaction)
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.ZRemRangeByScore(ctx, key, "-inf", cutoffStr)
				pipe.ZAdd(ctx, key, redis.Z{Score: score, Member: member})
				return nil
			})
			return err
		}, key)

		if err == nil {
			return true, nil
		}
		if errors.Is(err, errOverLimit) {
			return false, nil
		}
		if errors.Is(err, redis.TxFailedErr) {
			lastErr = err
			continue
		}
		return false, err
	}
	return false, lastErr
}

// Ensure OptimisticLimiter implements domain.RateLimiter at compile time.
var _ domain.RateLimiter = (*OptimisticLimiter)(nil)
