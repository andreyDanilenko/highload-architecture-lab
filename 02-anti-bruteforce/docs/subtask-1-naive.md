# 1. Naive Vault (In-Memory)

**What:** In-memory store of attempts with a mutex.  
**Why:** Show that in-memory rate limiting does not scale horizontally and has lock contention issues.

---

## Implementation steps

1. Create a `SlidingWindowStorage` (or equivalent) with `sync.RWMutex` and `map[string][]time.Time` (or map of IP → slice of timestamps).
2. Method `Allow(ip string, limit int, windowSec int64) bool`:
   - Take write lock.
   - Get list of timestamps for the IP.
   - Filter out timestamps outside the window (`time.Now().Unix() - windowSec`).
   - If length of remaining list < limit, append current timestamp and return `true`; else return `false`.
3. Endpoint `POST /login` (or `/resource/naive`):
   - Get IP from request context.
   - Call `storage.Allow(ip, 5, 60)`.
   - If `false` — return `429 Too Many Requests`.
   - If `true` — run "login logic" (e.g. sleep 50ms and return 200 OK).
4. Test: script that sends 10 sequential requests from one IP. The 6th request should get 429.

> **Implementation note:** when extracting the client IP from `http.Request`,
> strip the ephemeral port from `RemoteAddr` (use `net.SplitHostPort`) so that
> repeated requests from the same client are counted under the same IP. In
> production, real client IPs should come from trusted proxy headers
> (`X-Real-IP`, `X-Forwarded-For`) set by your load balancer / frontend, with
> any client-supplied values removed.

---

## What will be done

- In-memory store with mutex.
- Endpoint using this store.
- Proof of the problem: with two app instances behind a load balancer, the effective limit is doubled per instance.
