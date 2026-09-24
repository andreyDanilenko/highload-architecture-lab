# 3. Redis-Based Idempotency Provider

**What:** Replace the in-memory map with a Redis-backed `IdempotencyProvider` that can be shared across multiple instances.

**Why:** To share idempotency state across instances and investigate concurrency, persistence and recovery. Durability depends on Redis configuration and failure behavior; TTL is a retention policy, not a durability guarantee.

**Status:** Planned subtask. The current `RedisProvider` is a stub. Atomic Redis writes do not atomically include an external business effect.

---

## Implementation steps

1. **Redis-backed provider**
   - Implement `RedisProvider` with:
     - Fields: `client *redis.Client`, `logger *zap.Logger`, `keyPrefix string`, `lockTTL time.Duration`.
   - Methods:
     - `GetOrCreate(ctx, key, ttl)` using `SET NX`:
       - If key absent, create `pending` record with `LockID` and `LockExpiresAt`.
       - If present, deserialize JSON and return existing record.
     - `Complete(ctx, key, response)`:
       - Update `Status` to `completed`, store HTTP response payload, refresh TTL.
     - `Fail(ctx, key, err)`:
       - Update `Status` to `failed`, store error text, refresh TTL.
     - `Get(ctx, key)` and `Cleanup(ctx, olderThan)` as needed.

2. **Refactor endpoint**
   - Replace `InMemoryProvider` with `RedisProvider` in your wiring.
   - Ensure all instances of the service connect to the same Redis cluster.
   - Keep the idempotency logic local to the handler (middleware will come later).

3. **Behavior under concurrency**
   - Write a test:
     - Start multiple goroutines sending the same `Idempotency-Key`.
     - Verify one active owner while the lease is valid; other requests replay the result or return a conflict. Count business effects separately from handler executions.
   - Identify race windows in record transitions, then investigate owner checks in the next subtask.
   - Test application restart, Redis restart/failover, eviction and expiry. State which records can be lost in the tested configuration.
   - Crash after the business effect but before `Complete`; an absent result must not be treated as proof that no effect occurred.

---

## What will be done

- Implement `RedisProvider` that stores shared idempotency records in Redis with a defined retention window.
- Use this provider instead of the in-memory map for a target endpoint.
- Validate behavior under concurrent requests and across multiple instances.

