# 2. In-Memory Idempotency Store

**What:** Study an in-memory idempotency map with TTL in a single process. Deduplication lasts only while the relevant record and ownership remain valid; this does not establish exactly-once effects under failures.

**Why:** To introduce the idempotency pattern in the simplest possible implementation, and to surface its limitations (no persistence, no horizontal scaling). 

---

## Implementation steps

1. **In-memory store**
   - Define `IdempotencyRecord` with:
     - `Status` (`pending`, `completed`, `failed`).
     - `ResponseCode`, `ResponseBody`, `ResponseHeaders`.
     - `CreatedAt`, `UpdatedAt`.
   - Implement `InMemoryProvider`:
     - Fields: `mu sync.RWMutex`, `records map[string]*IdempotencyRecord`, `ttl time.Duration`.
     - Methods:
       - `GetOrCreate(ctx, key, ttl)`:
         - If key is absent — create a `pending` record atomically.
         - For expired records, first apply the retention and recovery policy: expiry does not prove that an earlier operation stopped or had no effect.
         - If exists — return existing.
       - `Complete/Fail` — update record fields; investigate how to reject completion from an old owner.
       - Optional `Cleanup` goroutine to remove stale entries.

2. **Handler integration (manual)**
   - For a single endpoint (e.g. `/payments`):
     - Read `Idempotency-Key` header.
     - If missing for POST/PUT — `400 Bad Request`.
     - Call `provider.GetOrCreate(key, ttl)`.
     - Behavior:
       - If new `pending` record: execute business logic, then `Complete`.
       - If `completed`: return stored response.
       - If `failed`: distinguish a confirmed failure without an effect from an unknown outcome before allowing a retry.

3. **Limitations**
   - Document clearly:
     - Data is lost on restart.
     - Multiple instances have disjoint maps — idempotency is **per instance**, not global.
     - TTL/cleanup limits retention, but a capacity bound needs an explicit policy.
     - The current provider has no owner token; an expired operation can finish after another request acquires the key.
     - A crash between the business effect and result storage can permit a duplicate effect after restart.

---

## What will be done

- Implement `InMemoryProvider` that satisfies a minimal `IdempotencyProvider` interface.
- Wrap one endpoint with manual idempotency logic using this provider.
- Document reliability and scaling limitations.

