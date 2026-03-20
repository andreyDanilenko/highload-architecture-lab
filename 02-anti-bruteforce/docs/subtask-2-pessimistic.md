# 2. Pessimistic Vault (Redis Lock)

**What:** Protect the endpoint using distributed locks in Redis.  
**Why:** Show the performance cost and lock contention of this approach.

---

## Implementation steps

1. Initialize Redis client.
2. Rate-limit logic in the service (not in Redis):
   - Lock key: `lock:rate:{ip}`.
   - Try to acquire lock: `SET lock:rate:{ip} <uuid> NX EX 1` (1 second TTL).
   - If failed — wait 50ms and retry (max 3 attempts); else return 500.
   - If lock acquired:
     - Read from Redis key `rate:{ip}` (Sorted Set).
     - Trim old entries (client-side).
     - Check count against limit.
     - If under limit, add new entry with `ZADD`.
     - Release lock (`DEL`).
   - On any error, ensure lock is released.
3. Endpoint `POST /resource/pessimistic`.
4. Test: script that sends 100 concurrent requests from one IP. Measure latency and compare with naive.

---

## What will be done

- Distributed lock mechanism.
- Latency measurement.
- Demonstration that locks are overkill for rate limiting.
