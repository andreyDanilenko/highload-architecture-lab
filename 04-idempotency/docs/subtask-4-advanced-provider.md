# 4. Production-Ready Idempotency (Middleware, Lua, Locks, Observability)

**What:** Upgrade the Redis-based provider to a production-grade idempotency system with atomic operations (Lua scripts), HTTP middleware, lock expiry handling, metrics, and tests.

**Why:** To guarantee strong correctness (no double execution), good UX (cached responses, clear 409 behavior), and full observability in production.

---

## Implementation steps

1. **HTTP middleware**
   - Implement `IdempotencyMiddleware`:
     - Extract `Idempotency-Key` from header for write methods (POST/PUT/PATCH/DELETE).
     - Validate key format (length, allowed chars).
     - Call `provider.GetOrCreate` and branch:
       - `pending` + created → execute handler once.
       - `pending` + existing → return `409 Conflict` + `Retry-After`.
       - `completed` → return cached response.
       - `failed` → decide: retry or return stored error.
   - Use `responseRecorder` to capture status code, headers, and body for storage.

2. **Lua scripts for atomicity and lock ownership**
   - Implement Redis Lua scripts:
     - `GetOrCreate` with lock:
       - Atomically create or update a `pending` record with `lock_id` and `lock_expires_at`.
       - Distinguish between "created", "locked by other", "completed", "failed".
     - `Complete/Fail` with lock check:
       - Ensure only the owner of the lock (`lock_id`) can finalize the record.
   - Wire these scripts into `RedisProvider` methods (`Eval`/`EvalSha`).

3. **Lock expiry and cleanup**
   - Background job that:
     - Scans keys with the idempotency prefix.
     - Uses Lua to detect `pending` records with expired `lock_expires_at` and releases locks safely.
   - Ensure stuck operations don't block idempotent retries forever.

4. **Metrics and logging**
   - Add `IdempotencyMetrics`:
     - Counters: total idempotent requests, cache hits, conflicts, errors.
     - Histogram: operation duration (get_or_create, complete, fail).
   - Structured logs (`IdempotencyLog`) containing:
     - Key, operation, status, duration, request ID, error.
   - Optional alerts in Prometheus:
     - High conflict rate.
     - High Redis error rate.

5. **Configuration and tests**
   - YAML-like config for:
     - Default TTL, per-endpoint overrides, Redis options, lock TTL, cleanup interval.
   - Tests:
     - Single and duplicate requests (cache behavior).
     - Concurrent requests with same key (only one handler execution).
     - Behavior on failure (failed status).
     - Basic benchmark with parallel idempotent requests.

---

## What will be done

- Introduce middleware that centralizes idempotency behavior for write endpoints.
- Make Redis operations fully atomic with Lua scripts and explicit lock ownership.
- Add lock expiry handling, metrics, and structured logging.
- Cover the provider with concurrency-focused tests and basic benchmarks.

