# 4. Atomic Vault (Redis Lua)

**What:** Atomic check and update via a Lua script on Redis.  
**Why:** Perform one check/update without command interleaving inside Redis, avoiding application-managed leases and WATCH conflict retries. Atomic execution alone does not establish durability or production readiness.

---

## Implementation steps

1. **Lua script** `sliding_window_log.lua`:
   - `KEYS[1]` — key for IP (e.g. `rate:{ip}`).
   - `ARGV[1]` — current timestamp (seconds or milliseconds).
   - `ARGV[2]` — window size in seconds.
   - `ARGV[3]` — max number of attempts (limit).
   - **Logic:**
     - `redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, ARGV[1] - ARGV[2])` — remove old entries.
     - `local current = redis.call('ZCARD', KEYS[1])` — current count.
     - If `current < tonumber(ARGV[3])`:
       - `redis.call('ZADD', KEYS[1], ARGV[1], ARGV[1])` — add new entry (member = timestamp).
       - `redis.call('EXPIRE', KEYS[1], ARGV[2])` — refresh TTL so key expires when idle.
       - Return allowed (e.g. `1`).
     - Else return blocked (e.g. `0`).
2. On app startup, load script with `SCRIPT LOAD` and store the SHA.
3. Service method `Allow(ip string, limit int, windowSec int64) bool`:
   - Call `EVALSHA` with stored SHA, keys, and args.
   - Interpret Redis response.
4. **Graceful degradation:** If Redis is down — fail-open (allow) or fail-close (500). For security, fail-close is simpler.
5. Endpoint `POST /vault/login`:
   - Extract IP.
   - Call `Allow(ip, 5, 60)` (5 attempts per IP per minute).
   - If `false` — 429.
   - If `true` — run "login" logic (e.g. mock password check).
6. Add Prometheus metrics: `rate_limit_total{ip, status}`.

---

## Load testing

These are target assertions for a fresh key, one observation window, healthy Redis and collision-free request members. The timestamp-only `ZADD` sketch above can merge concurrent attempts with equal timestamps; give each attempt a distinct member before expecting exact counts. Validate these conditions against the actual script.

1. **Scenario A (single IP):** 100 concurrent requests from one IP to `/vault/login`. Expect exactly 5× 200, 95× 429. Check that no more than five attempts are admitted in this controlled run.
2. **Scenario B (many IPs):** 100 concurrent requests from 100 different IPs. Expect 100× 200.
3. **Scenario C (two instances):** Run 2 app instances behind a load balancer. Repeat scenario A. Same result: 5 success, 95 blocked. Checks that both instances share this Redis-backed quota in this run; failover and clock behavior need separate tests.

---

## What will be done

- Atomic Lua script for sliding window.
- Script load and SHA caching.
- Protected login endpoint.
- Graceful degradation (e.g. fail-close when Redis down).
- Metrics integration.
- Load tests for correctness and performance.
