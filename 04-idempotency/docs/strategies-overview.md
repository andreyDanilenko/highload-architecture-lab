# Idempotency Strategies Overview

**Goal:** Compare strategies for preventing duplicate business effects under retries, timeouts, and failures; state each strategy’s retention window and failure assumptions. Without idempotency, repeated requests lead to double charges, duplicate orders, and broken business invariants.

---

## 1. No Idempotency (Baseline)

- **Flow:** Handler just executes the operation (e.g. charges a payment) every time it receives a request.
- **Behavior under retry:** Any client/network retry fully re-executes the operation (second payment, second order, etc.).
- **Use:** A baseline that exposes duplicate effects in this task. It provides no deduplication for the chosen non-idempotent operation.

---

## 2. In-Memory Idempotency

- **Flow:** In-memory `map[key]Result` protected by mutex. On each request:
  - Atomically create a `pending` record for a new key, then run the operation and store its result.
  - Replay a completed result or report an in-progress conflict for an existing key.
- **Pros:** Simple to implement, no external dependency.
- **Cons:** 
  - State is lost on restart.
  - Does not work with multiple instances (keys are local to process).
  - The current implementation has TTL/cleanup, but no capacity limit. Expiry and stale-owner behavior still need verification.
- **Use:** Demo of basic idempotent pattern and its limitations.

---

## 3. Redis-Based Idempotency Provider

- **Flow:** Central `IdempotencyProvider` backed by Redis:
  - `GetOrCreate(key)` creates a `pending` record using `SET NX` or returns existing one.
  - `Complete/Fail` updates record with response or error.
  - TTL controls how long records live.
- **State sharing:** Instances use the same Redis keyspace. Application restarts can retain records while Redis retains them; Redis restart/failover durability depends on configuration, and TTL/eviction can remove records.
- **Weaknesses:** 
  - Without careful ownership checks and atomic record transitions, concurrent requests can race.
  - A crash after an external effect but before recording the result leaves an unknown outcome. Lua alone does not close this gap.
- **Use:** A planned shared-state experiment; the repository’s Redis provider is currently a stub.

---

## 4. Advanced Idempotency (Ownership, Lua, Failure Handling)

- **Flow:** 
  - HTTP middleware extracts/validates idempotency key.
  - Provider:
    - Atomically creates/locks a record (Redis + Lua).
    - Distinguishes `pending`, `completed`, `failed` states.
  - First request executes handler and stores response.
  - Concurrent requests:
    - While `pending` — return `409 Conflict` + `Retry-After`.
    - After `completed` — return cached response.
- **Atomicity:** A Lua script can make Redis record transitions atomic and check an owner token. It does not include a separate database write, payment or email in that atomic step.
- **Reliability:** 
  - Expiry handling distinguishes unknown outcomes from confirmed failures; expiry does not stop the old executor.
  - Owner tokens protect the record. Coordinate the effect through a shared transaction or downstream idempotency, and reconcile unknown outcomes. Target-enforced fencing can reject stale-owner writes, but cannot deduplicate an effect already committed by a previous valid owner.
  - Separate result retention from the ownership lease; test restarts and takeover.
- **Observability:** Prometheus metrics and structured logs around idempotency operations.
- **Use:** A planned exercise in specifying and testing failure behavior. Lua, locks and metrics alone do not establish readiness for real payments or other critical operations.

---

## Summary table

| Strategy         | Storage       | Scope             | Reliability                 | Use Case                      |
|-----------------|---------------|-------------------|-----------------------------|-------------------------------|
| No Idempotency  | None          | Per request       | None                        | Anti-example only            |
| In-Memory       | Process RAM   | Single instance   | Lost on restart; expiry needs verification | Local experiments |
| Redis Provider  | Redis         | Shared keyspace   | Persistence-dependent; external effect gap | Planned shared-state experiment |
| Advanced        | Redis + Lua   | Shared keyspace   | Record atomicity; effect boundary must be addressed | Planned failure experiments |

