# Task 01: Atomic Inventory Counter

Implement a reservation system that handles thousands of concurrent requests without going negative or losing updates (race condition).

---

## Problem

Multiple threads or instances read the same stock, subtract, and write back — some deductions overwrite each other and are "lost":

```
Thread 1: Read stock = 5
Thread 2: Read stock = 5
Thread 1: Write stock = 4
Thread 2: Write stock = 4   ← one unit lost
```

Requirement: `stock_quantity` must never go negative.

---

## Task (overview)

Implement and compare four reservation strategies:

1. **Naive** — read-modify-write with no locking. Used only to demonstrate the race (100 requests → some updates lost). Do not use in production.
2. **Pessimistic** — row lock in the DB (`SELECT FOR UPDATE`) inside a single transaction. No lost updates; under high contention, slower.
3. **Optimistic** — `version` column; update only if version matches, retry on conflict. No long-held locks; faster than pessimistic when contention is low.
4. **Redis** — atomic counter in Redis (Lua), synced to PostgreSQL. Highest RPS; fits highload counters (stock, limits, views).

Where this appears in production: e-commerce (reserve on checkout), seat/table booking, bonus deduction, billing limits.

Step-by-step plans per subtask are in `docs/`:

- [docs/subtask-1-naive.md](docs/subtask-1-naive.md) — race demo
- [docs/subtask-2-pessimistic.md](docs/subtask-2-pessimistic.md) — SELECT FOR UPDATE
- [docs/subtask-3-optimistic.md](docs/subtask-3-optimistic.md) — version + retry
- [docs/subtask-4-redis.md](docs/subtask-4-redis.md) — Redis + PG

---

## Tech stack

- **Runtime:** Node.js 20+ (reference implementation in `implementation-node`)
- **Framework:** Fastify (routes, validation, error handler)
- **Language:** TypeScript (tsx for dev)
- **DB:** PostgreSQL 16+ — products, stock, version; transactions table (idempotency by `request_id`)
- **Cache/counter:** Redis 7+ — atomic ops (Lua) for the Redis strategy
- **Logging:** Pino (stdout + optional file)
- **Validation:** Zod (request body)
- **Infra:** Docker Compose (Postgres, Redis), DB reset and load-test scripts (bash, curl, jq)

---

## Running (shared for Node and Go)

Infrastructure lives in this folder (`01-atomic-inventory`): Docker Compose, init SQL, Makefile.

**1. Env (for Docker)**

```bash
cp env.example .env
# .env: DB_USER, DB_PASSWORD, DB_NAME, DB_PORT
```

**2. Start Postgres and Redis**

```bash
make infra-up
```

**3. Reset DB (stock 1000, clear transactions)**

```bash
make reset-db
```

**4. Run implementation**

- **Node:** `make run-node` or `cd implementation-node && make dev` — see [implementation-node/README.md](implementation-node/README.md).
- **Go:** placeholder; see [implementation-go/README.md](implementation-go/README.md).

Other: `make infra-down`, `make infra-logs`, `make infra-ps`, `make infra-reset`.

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

Details and how to read results: [scripts/README.md](scripts/README.md).

---

## Logs

The Node app logs with Pino: stdout and, when enabled, a file.

- **Files:** `implementation-node/logs/`. Filename: `run-<ISO-timestamp>.log` (created on startup when file logging is on).
- **File logging:** On by default when `NODE_ENV !== "test"` and `LOG_TO_FILE` is not `"0"`.
- **Level:** `LOG_LEVEL` (e.g. `info`, `debug`). Default `info`.
- **Disable file:** set `LOG_TO_FILE=0` when starting the app.

---

## Deliverables (checklist)

- All four strategies implemented (e.g. in Node under `implementation-node/`).
- Idempotency by `requestId` on every reserve endpoint.
- Load-test scripts: DB reset + race-test per strategy; naive shows lost updates; pessimistic, optimistic, redis show no loss.
- Docker Compose for Postgres and Redis; short docs in README and step-by-step plans in `docs/`.

Success criterion for correct strategies: 100 successful requests of 1 unit each → final stock exactly 100 less than initial; no negative stock in any test.
