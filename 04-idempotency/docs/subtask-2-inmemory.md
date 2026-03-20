# 2. In-Memory Idempotency Store

**What:** In-memory idempotency map with TTL, providing basic protection against duplicate execution for a single instance.

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
         - If key not in map or expired — create `pending` record and store.
         - If exists — return existing.
       - `Complete/Fail` — update record fields.
       - Optional `Cleanup` goroutine to remove stale entries.

2. **Handler integration (manual)**
   - For a single endpoint (e.g. `/payments`):
     - Read `Idempotency-Key` header.
     - If missing for POST/PUT — `400 Bad Request`.
     - Call `provider.GetOrCreate(key, ttl)`.
     - Behavior:
       - If new `pending` record: execute business logic, then `Complete`.
       - If `completed`: return stored response.
       - If `failed`: decide whether to repeat or return error.

3. **Limitations**
   - Document clearly:
     - Data is lost on restart.
     - Multiple instances have disjoint maps — idempotency is **per instance**, not global.
     - Memory can grow unbounded without proper cleanup.

---

## What will be done

- Implement `InMemoryProvider` that satisfies a minimal `IdempotencyProvider` interface.
- Wrap one endpoint with manual idempotency logic using this provider.
- Document reliability and scaling limitations.

