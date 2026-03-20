# Task 01: Atomic Inventory Counter

Reservation system under high concurrency: no negative stock, no lost updates.

---

## Problem

**Race condition (lost update):** multiple threads/instances read the same value, modify, and write back — last write overwrites others:

```
Thread 1: Read stock = 5
Thread 2: Read stock = 5
Thread 1: Write stock = 4
Thread 2: Write stock = 4   ← one unit lost
```

Requirement: `stock_quantity` must never go negative.

---

## Patterns and production context

| Pattern | What it solves | Where used in production |
|---------|----------------|--------------------------|
| **Pessimistic lock** | Serialize access via row lock (`SELECT FOR UPDATE`) | Strong consistency, medium load: checkout, booking |
| **Optimistic lock** | Avoid long-held locks; retry on version conflict | High load, low contention: inventory, versioned entities |
| **Atomic counter** | Single-key atomic ops (Redis Lua, DB `SET x = x - 1`) | High RPS: stock, views, rate limits |
| **Idempotency** | Same `requestId` → no double deduction | Retries, distributed systems |
| **Compensating transaction** | Rollback side effect if main op fails (e.g. Redis increment after PG fail) | Two-phase flows, eventual consistency |
| **Read/Write separation** | Hot path in Redis, PG as source of truth | CQRS-like: fast reads/writes, async sync to DB |

**Production domains:** e-commerce (reserve on checkout), seat/table booking, bonus deduction, billing limits.

---

## Task (overview)

Implement and compare four strategies:

1. **Naive** — read-modify-write, no locking. Demo only (shows lost updates).
2. **Pessimistic** — `SELECT FOR UPDATE` in transaction. Strong consistency; blocking under contention.
3. **Optimistic** — `version` column + retry on conflict. No long-held locks; good when conflicts are rare.
4. **Redis** — atomic Lua decrement, sync to PG by delta; compensating transaction on PG failure. Highest RPS.

Step-by-step plans in `docs/`:

- [docs/subtask-1-naive.md](docs/subtask-1-naive.md) — race demo
- [docs/subtask-2-pessimistic.md](docs/subtask-2-pessimistic.md) — SELECT FOR UPDATE
- [docs/subtask-3-optimistic.md](docs/subtask-3-optimistic.md) — version + retry
- [docs/subtask-4-redis.md](docs/subtask-4-redis.md) — Redis + PG (delta in PG, compensating transaction)
- [docs/strategies-overview.md](docs/strategies-overview.md) — comparison of all four strategies

---

## Tech stack

- **DB:** PostgreSQL 18+ — products, stock, version; transactions (idempotency by `request_id`)
- **Cache:** Redis 8+ — atomic Lua for Redis strategy
- **Infra:** Docker Compose, Makefile, load-test scripts (bash, curl, jq)
- **Reference:** Node.js 24+, Fastify, TypeScript — see [node/README.md](node/README.md); Go 1.25+ — see [go/README.md](go/README.md)

---

## Quick start

```bash
cp env.example .env
make infra-up
make reset-db
make run-node
```

Other: `make infra-down`, `make infra-logs`, `make infra-reset`.

---

## Database schema

```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) UNIQUE NOT NULL,
    stock_quantity INT NOT NULL CHECK (stock_quantity >= 0),
    version INT DEFAULT 0
);

CREATE TABLE inventory_transactions (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    request_id VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);
```

Seed: one product `SKU-TEST-001` with stock 1000.

---

## API

- `GET /api/v1/inventory/stock/:sku` — current stock (from PG).
- `POST /api/v1/inventory/reserve` — naive (tests only).
- `POST /api/v1/inventory/reserve/pessimistic` — transaction lock.
- `POST /api/v1/inventory/reserve/optimistic` — version + retry.
- `POST /api/v1/inventory/reserve/redis` — Redis atomic counter.

Reserve body: `{ "sku": "SKU-TEST-001", "quantity": 1, "requestId": "uuid" }`.  
Responses: 200 — success; 409 — insufficient stock; 404 — product not found; 422 — invalid body.

---

## Testing

**Unit:** Idempotency (same `requestId` does not create a second deduction); concurrent deductions never go negative; 409 when stock is insufficient.

**Load:** Scripts in `scripts/load-test/`. Before each run, reset DB: `./scripts/reset-db.sh`. Then e.g.:

- `./scripts/load-test/race-test-naive.sh` — expect lost updates (actual stock > initial - success).
- `./scripts/load-test/race-test-pessimistic.sh`, `race-test-optimistic.sh`, `race-test-redis.sh` — expect consistency (actual = initial - success).

Details: [scripts/README.md](scripts/README.md).

---

## Problems and limitations per strategy

| Strategy    | Main problems / risks |
|------------|------------------------|
| **Naive**  | Lost updates under concurrency (read–modify–write without lock). Risk of oversell. Use only for demos/load-test, not production. |
| **Pessimistic** | High load on DB: row locks block other transactions; under contention, latency grows. Deadlocks possible if order of locking differs across requests. No lost updates; strong consistency. |
| **Optimistic** | Retries under contention (version conflict) increase latency; many concurrent updates to same SKU can exhaust `maxOptimisticRetries`. No long-held locks; good when conflict rate is low. |
| **Redis**  | Two sources of truth (Redis + PG): if PG write fails after Redis decrement, Redis stays ahead and becomes inconsistent. Requires compensating transaction (rollback in Redis) and/or reconciliation. If Redis is down, reserve path fails unless fallback to another strategy. |

For Redis: compensating transaction, drift risks, and **stabilization plan** (no code) — [docs/redis-stabilization-plan.md](docs/redis-stabilization-plan.md).

---

## Deliverables

- [ ] All four strategies implemented (e.g. in `node/`)
- [ ] Idempotency by `requestId` on every reserve endpoint
- [ ] Load-test scripts: naive shows lost updates; pessimistic, optimistic, redis show consistency
- [ ] Docker Compose, step-by-step plans in `docs/`

**Success criterion:** 100 requests × 1 unit → final stock = initial − 100; no negative stock.

---

## README template (for reuse)

Structure for similar task READMEs:

1. **Title + one-liner** — what the task does
2. **Problem** — core issue (with example if helpful)
3. **Patterns and production context** — design patterns, what each solves, where used in prod
4. **Task (overview)** — strategies/approaches, links to detailed docs
5. **Tech stack** — tools (DB, cache, infra), reference impl link
6. **Quick start** — minimal commands
7. **Schema / API** — contracts
8. **Testing** — unit + load, success criteria
9. **Limitations** — trade-offs per approach
10. **Deliverables** — checklist + success criterion
