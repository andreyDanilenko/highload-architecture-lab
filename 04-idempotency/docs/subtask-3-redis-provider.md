# 3. Redis-Based Idempotency Provider

**What:** Replace the in-memory map with a Redis-backed `IdempotencyProvider` that can be shared across multiple instances.

**Why:** To make idempotency state durable (within TTL) and correct in a horizontally scaled environment.

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
     - Ensure only one actual business operation runs; others read the stored record or fail fast.
   - At this stage, atomicity may still be imperfect (race windows with `GET` + `SET`), which will be addressed in the next subtask.

---

## What will be done

- Implement `RedisProvider` that persists idempotency records in Redis with TTL.
- Use this provider instead of the in-memory map for a target endpoint.
- Validate behavior under concurrent requests and across multiple instances.

