# Task 04: Idempotency Key Provider

Guarantee **exactly-once effect** for side-effectful HTTP operations (payments, orders, emails) in the presence of retries, timeouts, and network errors.

---

## Problem

Clients, gateways, and load balancers retry requests. Users double-click buttons. Networks drop responses after the server has already executed the operation. Without idempotency, the same request can be processed multiple times, leading to **double charges**, **duplicate orders**, and broken business invariants.

Idempotency solves this by introducing an **Idempotency-Key**: repeated requests with the same key return the **same** outcome without repeating the side effect.

---

## Task (overview)

Implement and compare four strategies:

1. **No Idempotency** — execute operation on every request. Demo-only baseline to show the bug.
2. **In-Memory** — store `key → result` in a local map with TTL; works only within a single process.
3. **Redis Provider** — store idempotency records in Redis with TTL so it works across instances and survives restarts.
4. **Advanced / Production** — middleware + atomic Redis operations (Lua) + explicit locks/ownership, lock expiry handling, metrics, and concurrency-focused tests.

Step-by-step plans per subtask are in `docs/`:

- [docs/subtask-1-naive.md](docs/subtask-1-naive.md) — baseline: no idempotency
- [docs/subtask-2-inmemory.md](docs/subtask-2-inmemory.md) — in-memory provider with TTL
- [docs/subtask-3-redis-provider.md](docs/subtask-3-redis-provider.md) — Redis-backed provider (shared, durable)
- [docs/subtask-4-advanced-provider.md](docs/subtask-4-advanced-provider.md) — production-ready (middleware + Lua + locks + metrics)
- [docs/strategies-overview.md](docs/strategies-overview.md) — comparison of all four strategies
- [docs/spec/specification.ru.md](docs/spec/specification.ru.md) — implementation guide (best practices, RU)

---

## API (expected)

This task is about the **idempotency layer**, so the concrete domain can be “payments” / “orders” / “emails”. A typical shape:

- `POST /payments` — side-effectful endpoint protected by idempotency.

Expected idempotency behaviors:

- **Missing key** (write methods): `400 Bad Request`
- **First request with a new key**: execute business operation → `200 OK` (or `201`) and persist response under that key
- **Repeat request with same key after completion**: return cached response (same status, headers, body)
- **Concurrent request while first is `pending`**: `409 Conflict` + `Retry-After` (or similar)

Notes:
- Prefer returning `409` for “still processing” (pending) rather than waiting indefinitely.
- For overload/backpressure, `429` can also be used (not required by default).

---

## What to verify

- **Correctness under retries**: sending the same request (same `Idempotency-Key`) multiple times results in **one** side effect.
- **Concurrency safety**: N parallel requests with the same key do not execute business logic more than once.
- **Cache semantics**: completed responses are replayed exactly (status, headers, body).
- **TTL & cleanup**: old keys expire; memory/Redis usage is bounded.
- **Lock expiry** (advanced): stuck `pending` operations do not block forever.
- **Observability** (advanced): metrics/logs show cache hits, conflicts, errors, latencies.

---

## Problems and limitations per strategy

| Strategy | Main problems / risks |
|----------|------------------------|
| **No Idempotency** | Repeated requests repeat side effects (double charge). Demo only. |
| **In-Memory** | Lost on restart; does not work across instances; needs TTL/cleanup to avoid leaks. |
| **Redis Provider** | Shared/durable, but can still have races without strict atomicity/locking. |
| **Advanced / Prod** | More complexity: Lua scripts, lock ownership, expiry policy, and testing. |

