# Idempotency Strategies Overview

**Goal:** Guarantee **exactly-once** effect for operations (payments, orders, emails) in the presence of retries, timeouts, and network errors. Without idempotency, repeated requests lead to double charges, duplicate orders, and broken business invariants.

---

## 1. No Idempotency (Baseline)

- **Flow:** Handler just executes the operation (e.g. charges a payment) every time it receives a request.
- **Behavior under retry:** Any client/network retry fully re-executes the operation (second payment, second order, etc.).
- **Use:** Only as a baseline example to show why idempotency is required; **never** acceptable in production.

---

## 2. Naive In-Memory Idempotency

- **Flow:** In-memory `map[key]Result` protected by mutex. On each request:
  - If key not present — run operation, store result, return it.
  - If key present — immediately return stored result.
- **Pros:** Simple to implement, no external dependency.
- **Cons:** 
  - State is lost on restart.
  - Does not work with multiple instances (keys are local to process).
  - Memory grows without bounds; no TTL or cleanup.
- **Use:** Demo of basic idempotent pattern and its limitations.

---

## 3. Redis-Based Idempotency Provider

- **Flow:** Central `IdempotencyProvider` backed by Redis:
  - `GetOrCreate(key)` creates a `pending` record using `SET NX` or returns existing one.
  - `Complete/Fail` updates record with response or error.
  - TTL controls how long records live.
- **Consistency:** Redis gives a single shared store across instances; records survive process restarts.
- **Weaknesses:** 
  - Without careful locking and atomic operations, concurrent requests with the same key can still race.
- **Use:** Foundation for production idempotency, but needs stronger atomicity.

---

## 4. Production-Ready Idempotency (Locks, Lua, Middleware)

- **Flow:** 
  - HTTP middleware extracts/validates idempotency key.
  - Provider:
    - Atomically creates/locks a record (Redis + Lua).
    - Distinguishes `pending`, `completed`, `failed` states.
  - First request executes handler and stores response.
  - Concurrent requests:
    - While `pending` — return `409 Conflict` + `Retry-After`.
    - After `completed` — return cached response.
- **Atomicity:** Lua scripts ensure read-modify-write is atomic and enforces lock ownership.
- **Reliability:** 
  - Lock expiry handling for stuck operations.
  - TTLs and cleanup jobs.
- **Observability:** Prometheus metrics and structured logs around idempotency operations.
- **Use:** Recommended approach for payments, order creation, and any critical side-effectful API.

---

## Summary table

| Strategy         | Storage       | Scope             | Reliability                 | Use Case                      |
|-----------------|---------------|-------------------|-----------------------------|-------------------------------|
| No Idempotency  | None          | Per request       | None                        | Anti-example only            |
| In-Memory       | Process RAM   | Single instance   | Lost on restart / no TTL    | Demo / local experiments     |
| Redis Provider  | Redis         | All instances     | Durable, but races possible | Baseline shared idempotency  |
| Prod-Ready      | Redis + Lua   | All instances     | Atomic, locked, observable  | Real payments / orders       |

