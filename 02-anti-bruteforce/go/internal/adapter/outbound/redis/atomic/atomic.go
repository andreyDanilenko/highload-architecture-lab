package atomic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"anti-bruteforce/internal/domain"

	"github.com/redis/go-redis/v9"
)

// AtomicLimiter implements sliding-window rate limiting using a Redis Lua script.
// Key idea: check + record is done atomically inside Redis, so no locks and no retries.
type AtomicLimiter struct {
	client *redis.Client

	mu        sync.Mutex
	scriptSHA string
}

func NewAtomicLimiter(client *redis.Client) *AtomicLimiter {
	return &AtomicLimiter{client: client}
}

// Lua script:
// - KEYS[1] = key for IP (e.g. rate:{ip})
// - ARGV[1] = nowSec (unix seconds)
// - ARGV[2] = windowSec
// - ARGV[3] = limit
// - ARGV[4] = unique member (to avoid collisions inside same second)
//
// Returns:
// - 1 if allowed
// - 0 if blocked
const slidingWindowLogLua = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]

-- Keep entries in [now-window, now) by removing strictly-less-than cutoff.
local cutoff = now - window
redis.call('ZREMRANGEBYSCORE', key, 0, cutoff - 1)

local current = redis.call('ZCARD', key)
if current < limit then
  redis.call('ZADD', key, now, member)
  -- TTL with buffer: if key is idle, it disappears.
  redis.call('EXPIRE', key, window * 2)
  return 1
end

return 0
`

func (a *AtomicLimiter) ensureScriptLoaded(ctx context.Context) (string, error) {
	// Fast path: already loaded
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.scriptSHA != "" {
		return a.scriptSHA, nil
	}

	sha, err := a.client.ScriptLoad(ctx, slidingWindowLogLua).Result()
	if err != nil {
		return "", err
	}
	a.scriptSHA = sha
	return sha, nil
}

func (a *AtomicLimiter) Allow(ip string, limit int, windowSec int64) (allowed bool, err error) {
	if limit <= 0 || windowSec <= 0 {
		return false, nil
	}

	ctx := context.Background()

	sha, err := a.ensureScriptLoaded(ctx)
	if err != nil {
		// Fail-close for security: when storage fails, treat as blocked.
		return false, nil
	}

	key := fmt.Sprintf("rate:%s", ip)
	now := time.Now()
	nowSec := now.Unix()
	member := fmt.Sprintf("%d", now.UnixNano())

	// EVALSHA returns whatever the Lua returns: 1 allowed / 0 blocked
	blockedOrAllowed, err := a.client.
		EvalSha(ctx, sha, []string{key}, nowSec, windowSec, limit, member).
		Int()
	if err != nil {
		// Fail-close for security
		return false, nil
	}

	return blockedOrAllowed == 1, nil
}

var _ domain.RateLimiter = (*AtomicLimiter)(nil)

