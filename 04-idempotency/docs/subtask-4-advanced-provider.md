# 4. Advanced Idempotency (Middleware, Lua, Ownership, Observability)

**What:** Extend the planned Redis provider with atomic record transitions (Lua scripts), HTTP middleware, ownership checks, expiry handling, metrics, and failure tests.

**Why:** To specify and test the boundaries of duplicate-effect protection and recovery. These mechanisms do not by themselves establish production readiness or exactly-once execution.

**Status:** Planned work. A Redis owner token protects the idempotency record, not an external payment or database write. Coordinate the record and effect through a shared transaction or downstream idempotency, and reconcile unknown outcomes before an unprotected retry. Target-enforced fencing adds protection against stale-owner writes; it does not deduplicate an effect already committed by a previous valid owner.

---

## Implementation steps

1. **HTTP middleware**
   - Implement `IdempotencyMiddleware`:
     - Extract `Idempotency-Key` from header for write methods (POST/PUT/PATCH/DELETE).
     - Validate key format, scope it to the caller/operation, and reject reuse with a different request payload.
     - Call `provider.GetOrCreate` and branch:
       - `pending` + created → admit the current owner to execute the handler.
       - `pending` + existing → return `409 Conflict` + `Retry-After`.
       - `completed` → return cached response.
       - `failed` → retry only when the effect is known to be absent or protected against duplication; otherwise reconcile the unknown outcome.
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
     - Detects `pending` records with expired `lock_expires_at` for recovery. Expiry does not stop the old executor or undo its effect.
   - Define when takeover is safe, reject stale record updates, and prevent duplicate effects at the target. Retain unknown outcomes for reconciliation instead of blindly retrying.

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
     - Concurrent requests with the same key during a valid lease (one active owner).
     - Paused owner resuming after expiry and takeover; count committed effects.
     - Crash between the effect and result storage, including an error from `Complete`.
     - Confirmed failures versus unknown outcomes, and retries after record expiry.
     - Basic benchmark with parallel idempotent requests.

---

## What will be done

- Introduce middleware that centralizes idempotency behavior for write endpoints.
- Make the required Redis record transitions atomic with Lua scripts and explicit lock ownership; document the separate boundary for business effects.
- Add lock expiry handling, metrics, and structured logging.
- Cover the provider with concurrency-focused tests and basic benchmarks.

