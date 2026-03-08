# Load & Race Condition Tests

Scripts to compare **naive** vs **pessimistic** (and later optimistic/Redis) reserve strategies under concurrent load.

---

## Prerequisites

- **Server** running (e.g. `cd implementation-node && make infra-up && make dev`)
- **Database** with seed data (stock 1000 for SKU-TEST-001). Use `reset-db.sh` before each run if you ran tests before.
- **Tools:** `curl`, `jq` (macOS: `brew install jq`)

---

## Scripts

| Script | Endpoint | Purpose |
|--------|----------|--------|
| `race-test.sh` | Configurable | Generic: pass base URL and reserve path. |
| `race-test-naive.sh` | `POST /api/v1/inventory/reserve` | Naive update (no locking) — **expect lost updates**. |
| `race-test-pessimistic.sh` | `POST /api/v1/inventory/reserve/pessimistic` | Pessimistic (SELECT FOR UPDATE) — **expect no lost updates**. |
| `race-test-optimistic.sh` | `POST /api/v1/inventory/reserve/optimistic` | Optimistic (version + retry) — **expect no lost updates**. |
| `race-test-redis.sh` | `POST /api/v1/inventory/reserve/redis` | Redis atomic counter + PG persistence — **expect no lost updates**. Clears Redis key so first request seeds from PG. |

---

## Usage

### 1. Reset database (recommended before each test run)

From repo root or `01-atomic-inventory`:

```bash
./scripts/reset-db.sh
```

Requires Docker container `inventory-postgres` and DB `inventory`. Restores stock to 1000 and clears transactions.

### 2. Run naive test (demonstrate race condition)

```bash
cd 01-atomic-inventory/scripts/load-test
chmod +x race-test.sh race-test-naive.sh race-test-pessimistic.sh ./race-test-optimistic.sh ./race-test-redis.sh

./race-test-naive.sh
# or: ./race-test-naive.sh http://localhost:3000
```

**Expected:** Many 200 responses, but **final stock > (initial - success)** due to lost updates. Example: 100 success, but stock dropped by less than 100.

### 3. Run pessimistic test (no race condition)

```bash
./scripts/reset-db.sh   # reset again
./race-test-pessimistic.sh
```

**Expected:** Final stock = initial − number of successful reserves. No lost updates.

### 4. Run optimistic test

```bash
./scripts/reset-db.sh
./race-test-optimistic.sh
```

**Expected:** No lost updates (same as pessimistic).

### 5. Run Redis test

Reset DB, then run the Redis test. The script clears the Redis key `inventory:stock:SKU-TEST-001` so the first request seeds stock from PostgreSQL.

**Important:** If you changed the Node server code (e.g. atomic Lua for Redis reserve), **restart the dev server** (`npm run dev` or `make dev`) so it loads the new code. Otherwise the test may still show 0 successful requests.

```bash
./scripts/reset-db.sh
./race-test-redis.sh
```

**Expected:** No lost updates (e.g. 100 successful, final stock = initial − 100). The script clears the Redis key before the test: it uses `redis-cli` if in PATH, or `docker exec inventory-redis redis-cli DEL ...` if the Redis container is running (typical when using `make infra-up`). If the key is never cleared (no redis-cli and no Docker), the key may be 0 from a previous run and all requests will fail with "insufficient stock".

### 6. Generic script (custom URL and path)

```bash
./race-test.sh [BASE_URL] [RESERVE_PATH]
```

Examples:

```bash
./race-test.sh
# → POST http://localhost:3000/api/v1/inventory/reserve (naive)

./race-test.sh http://localhost:3000 /api/v1/inventory/reserve/pessimistic
# → pessimistic

./race-test.sh http://localhost:8080 /api/v1/inventory/reserve
# → different host (e.g. Go impl on 8080)
```

---

## Quick start (full flow)

```bash
# 1. Reset DB
./scripts/reset-db.sh

# 2. Naive — see race condition
./scripts/load-test/race-test-naive.sh

# 3. Reset again
./scripts/reset-db.sh

# 4. Pessimistic — no lost updates
./scripts/load-test/race-test-pessimistic.sh
```

---

## Interpreting results

- **Successful** — count of HTTP 200 responses.
- **Failed** — non-200 (e.g. 409 insufficient stock when stock is exhausted).
- **Expected stock** — `initial - successful` (how much should remain if every success reserved 1).
- **Actual stock** — value from `GET /api/v1/inventory/stock/SKU-TEST-001`.

If **actual < expected** (more stock left than expected), some updates were lost (race).  
If **actual === expected**, no lost updates.
