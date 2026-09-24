# Anti-Bruteforce Strategies Overview

**Goal:** Protect a critical endpoint (e.g. `/login`) from password bruteforce using a **sliding window** (Sliding Window Log) rate limit. Why not simple TTL? Because "reset counter after N seconds" is vulnerable: an attacker can send N requests in the last second of the window and N in the first second of the next, effectively doubling the limit. The sliding window counts requests in the last N seconds at any moment.

---

## 1. Naive (In-Memory Store)

- **Flow:** Store in process memory (Map) per-IP lists of timestamps for attempts. On each request, filter timestamps inside the window (now - windowSize). If count exceeds limit — block.
- **Storage:** Process memory (e.g. Map with cleanup of old entries).
- **Concurrency:** This implementation holds a mutex around check/update, so it does not demonstrate an unprotected map race. High concurrency can create contention; unused IP keys also need cleanup.
- **Scaling:** With 2 app instances, the limit is enforced per instance. An attacker can spread load and get 2× the effective limit.
- **Use:** Demonstrate local quotas and mutex contention. Local limiting can protect one instance, but cannot alone enforce a cluster-wide quota.

---

## 2. Pessimistic (Redis Locks)

- **Flow:** Before updating attempt list for an IP, acquire a lock (e.g. `SET key NX EX`). If acquired, read list, update, save, release lock. If not — wait and retry.
- **Storage:** Redis.
- **Concurrency:** The lease serializes cooperating callers only while it remains valid. Test a paused holder exceeding TTL, ownership-safe release and Redis failover.
- **Cost:** High overhead. Lock acquire/release per request adds RTT to Redis and latency. Under contention on one IP, "herd blocking" — many requests waiting for the same lock.
- **Use:** When contention is low and simplicity matters; inefficient for high load.

---

## 3. Optimistic (Redis WATCH)

- **Flow:** Redis transaction with `WATCH`. `WATCH key:ip`, read current list. If over limit — abort. Else `MULTI`; `ZADD` new timestamp; `EXEC`. If `EXEC` returns nil (key changed), retry from the start.
- **Storage:** Redis (Sorted Set, score = timestamp).
- **Race:** Optimistic locking; one of two concurrent updaters retries.
- **Cost:** Under high conflict (same IP, many requests) — many retries, more Redis load. WATCH has its own overhead.
- **Use:** When conflicts are rare (e.g. different IPs). Not ideal for bruteforce protection, where the attacker creates high contention.

---

## 4. Atomic (Redis Lua)

- **Flow:** All logic for one IP runs inside one Lua script on Redis. Client sends `EVALSHA` with (IP key, limit, window size, current timestamp). Script: remove old entries from Sorted Set, count remaining, if count < limit then ZADD new entry and return allowed, else return blocked.
- **Storage:** Redis (Sorted Set).
- **Concurrency:** Redis executes the script without interleaving other commands. It avoids an application-managed lock and WATCH conflict retries; it does not guarantee persistence or resolve an unknown result after a network timeout.
- **Cost:** One successful script invocation; measure server execution time, network latency and recovery after script-cache misses.
- **Use:** A common Redis rate-limiting approach. Correct counting, bounded keys, time semantics and failure policy still need validation.

---

## Summary table

| Strategy    | Storage        | Concurrency              | Scaling        | Main trade-off                |
|-------------|----------------|--------------------------|----------------|-------------------------------|
| Naive       | In-Memory Map  | Mutex around check/update | Per-instance quota | Contention and key cleanup |
| Pessimistic | Redis (Lock)   | Serialized via locks     | Yes, with delay| High latency from locks      |
| Optimistic  | Redis (WATCH)  | Retry on conflict        | Yes, with load | Many retries under contention|
| **Atomic**  | **Redis (Lua)**| **Atomic script execution** | **Shared Redis state** | **Failover, counting and failure policy** |
