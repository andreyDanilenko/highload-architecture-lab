# 3. Optimistic Vault (Redis WATCH)

**What:** Use Redis transactions (WATCH / MULTI / EXEC) to avoid races without locks.  
**Why:** Show an alternative to locks and the retry load under contention.

---

## Implementation steps

1. Function `attempt(ip string, limit int, windowSec int64) bool`:
   - Start transaction: `WATCH rate:{ip}`.
   - Get current entries (`ZRANGE ... WITHSCORES`).
   - Trim old entries client-side and count remaining.
   - If count >= limit: `UNWATCH` and return `false`.
   - Else: `MULTI`, `ZADD` new timestamp, `EXEC`.
   - If `EXEC` returns nil (key was modified), retry (call `attempt` again in a loop).
2. Wrap `attempt` in a loop with a max retry count (e.g. 3).
3. Endpoint `POST /resource/optimistic`.
4. Test: 100 concurrent requests from one IP. Observe logs for conflict and retry count.

---

## What will be done

- Optimistic locking with WATCH.
- Retry mechanism.
- High Redis load from retries when many requests hit the same key.
