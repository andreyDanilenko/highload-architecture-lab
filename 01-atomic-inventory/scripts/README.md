# Load & Race Condition Tests

Scripts to compare reserve strategies under load: naive (no locking), pessimistic, optimistic, Redis.

**Prerequisites:** running server (Node/Go), Postgres and Redis (`make infra-up` from `01-atomic-inventory`), `curl` and `jq` (`brew install jq`).

---

## Scripts

- **`reset-db.sh`** — reset DB: stock 1000, clear transactions. Run from `01-atomic-inventory/scripts` or repo root; container `inventory-postgres` must be running.
- **`race-test.sh [BASE_URL] [RESERVE_PATH]`** — generic: 100 concurrent POSTs to the given URL.
- **`race-test-naive.sh`** — `POST …/reserve` (naive). Expect lost updates (actual stock > initial − success).
- **`race-test-pessimistic.sh`** — `POST …/reserve/pessimistic`. Expect no lost updates.
- **`race-test-optimistic.sh`** — `POST …/reserve/optimistic`. Expect no lost updates (some requests may 409 on version conflict).
- **`race-test-redis.sh`** — `POST …/reserve/redis`. Clears Redis key before test so stock is seeded from PG. Expect no lost updates.

---

## Quick start (copy and run)

From `01-atomic-inventory`:

```bash
make infra-up
make reset-db
cd node && npm run dev
```

In another terminal, from `01-atomic-inventory`:

```bash
./scripts/reset-db.sh
./scripts/load-test/race-test-naive.sh
```

Expect: many 200s, but **actual stock > initial − successful** (lost updates). Then:

```bash
./scripts/reset-db.sh
./scripts/load-test/race-test-pessimistic.sh
```

Expect: **actual stock === initial − successful** (no lost updates).

---

## Example commands

Reset DB (from `01-atomic-inventory` or from `scripts`):

```bash
./scripts/reset-db.sh
```

Naive / pessimistic / optimistic / Redis (from `01-atomic-inventory`):

```bash
./scripts/load-test/race-test-naive.sh
./scripts/load-test/race-test-pessimistic.sh
./scripts/load-test/race-test-optimistic.sh
./scripts/load-test/race-test-redis.sh
```

Custom URL and path (e.g. different host or port):

```bash
./scripts/load-test/race-test.sh http://localhost:3000 /api/v1/inventory/reserve
./scripts/load-test/race-test.sh http://localhost:8080 /api/v1/inventory/reserve/pessimistic
```

After changing server code, restart `npm run dev` before the Redis test. The script clears key `inventory:stock:SKU-TEST-001` via `redis-cli` or `docker exec inventory-redis redis-cli DEL ...`.

---

## Reading the output

- **Successful** — count of 200 responses.
- **Failed** — non-200 (e.g. 409 when stock is exhausted).
- **Expected stock** = initial − successful.
- **Actual stock** — from `GET …/stock/SKU-TEST-001`.

If **actual > expected** — some reserves were lost (race). If **actual === expected** — no race.
