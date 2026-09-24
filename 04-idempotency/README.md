# Task 04: Idempotency Key Provider

Study how to prevent duplicate business effects in side-effectful HTTP operations (payments, orders, emails), and identify the limits of each strategy under retries, timeouts, and failures.

---

## Problem

Clients, gateways, and load balancers retry requests. Users double-click buttons. Networks drop responses after the server has already executed the operation. Without idempotency, the same request can be processed multiple times, leading to **double charges**, **duplicate orders**, and broken business invariants.

An **Idempotency-Key** identifies one logical operation across retries. Within a defined retention window, matching requests can replay a stored result. Preventing duplicate business effects also requires coordinating that record with the effect: a Redis record alone does not make an external payment or database write atomic. A timeout can leave the outcome unknown.

---

## Current implementation

The Go implementation includes `noop` and `memory` providers. `RedisProvider` is a stub that returns `ErrProviderNotAvailable`; Redis and advanced behavior below are planned work. The memory provider has TTL/cleanup, but expiry, ownership and recovery need further verification. The payment operation is simulated, and no production guarantees have been established.

## Task (overview)

Implement and compare four strategies:

1. **No Idempotency** — execute operation on every request. Demo-only baseline to show the bug.
2. **In-Memory** — store `key → result` in a local map with TTL; works only within a single process.
3. **Redis Provider** — share idempotency records across instances. Application restarts can retain records while Redis retains them; Redis restart/failover behavior depends on persistence, replication and eviction settings.
4. **Advanced / Failure Handling** — middleware + atomic Redis operations (Lua) + explicit locks/ownership, lock expiry handling, metrics, and concurrency-focused tests.

Step-by-step plans per subtask are in `docs/`:

- [docs/subtask-1-naive.md](docs/subtask-1-naive.md) — baseline: no idempotency
- [docs/subtask-2-inmemory.md](docs/subtask-2-inmemory.md) — in-memory provider with TTL
- [docs/subtask-3-redis-provider.md](docs/subtask-3-redis-provider.md) — Redis-backed provider (shared state and persistence experiments)
- [docs/subtask-4-advanced-provider.md](docs/subtask-4-advanced-provider.md) — advanced experiments (middleware + Lua + ownership + metrics)
- [docs/strategies-overview.md](docs/strategies-overview.md) — comparison of all four strategies
- [docs/spec/specification.ru.md](docs/spec/specification.ru.md) — design guide with illustrative code sketches (RU)

---

## API (expected)

This task is about the **idempotency layer**, so the concrete domain can be “payments” / “orders” / “emails”. A typical shape:

- `POST /payments` — side-effectful endpoint protected by idempotency.

Expected idempotency behaviors:

- **Missing key** (write methods): `400 Bad Request`
- **First request with a new key**: execute business operation → `200 OK` (or `201`) and persist response under that key
- **Repeat request with the same key and payload after completion, within retention**: replay the stored response according to the endpoint contract
- **Same key with a different payload or operation scope**: reject the mismatch
- **Concurrent request while first is `pending`**: `409 Conflict` + `Retry-After` (or similar)

Notes:
- Prefer returning `409` for “still processing” (pending) rather than waiting indefinitely.
- For overload/backpressure, `429` can also be used (not required by default).

---

## What to verify

- **Correctness under retries**: count committed effects independently of HTTP responses; test retries within and after retention, and failure between the effect and result storage.
- **Concurrency safety**: within a valid ownership period, concurrent matching requests admit one active owner. Test a paused owner resuming after expiry; one handler execution and one committed business effect are different properties.
- **Cache semantics**: define and verify which status, headers and body are stored and replayed. This is expected behavior, not a claim about the current implementation.
- **TTL & cleanup**: verify expiry and measure storage under the stated request rate. TTL limits retention, not total memory by itself; expiry also ends the deduplication window.
- **Lock expiry** (advanced): distinguish a lost owner from a slow one. Define recovery for unknown outcomes and protect the effect against stale owners before permitting takeover.
- **Observability** (advanced): metrics/logs show cache hits, conflicts, errors, latencies.

---

## Problems and limitations per strategy

| Strategy | Main problems / risks |
|----------|------------------------|
| **No Idempotency** | Repeated requests repeat side effects (double charge). Demo only. |
| **In-Memory** | Lost on restart; does not work across instances; needs TTL/cleanup to avoid leaks. |
| **Redis Provider** | Shared state; durability depends on configuration. Atomic Redis operations do not include external effects. |
| **Advanced** | Ownership and failure handling add complexity; correctness remains scoped to the tested failure model and effect boundary. |

