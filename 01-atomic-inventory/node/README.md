# implementation-node

Node.js (Fastify + TypeScript) implementation with four reserve strategies: naive, pessimistic, optimistic, Redis.

Infrastructure (Docker, Postgres, Redis, DB reset) is shared and run from the parent folder — see [Running](#running) below.

---

## Prerequisites

- Node.js 20+
- npm
- Docker & Docker Compose (for Postgres and Redis; started from `01-atomic-inventory`)
- Optional: `curl`, `jq` for load tests

---

## Quick start

```bash
npm install
cp env.example.sh .env   # or copy key=value lines into .env
```

From **01-atomic-inventory** (parent folder) start infra and reset DB, then run the app:

```bash
cd ../
make infra-up
make reset-db
cd implementation-node && make dev
```

Or in one go from `01-atomic-inventory`:

```bash
make infra-up
make reset-db
make run-node
```

Server: **http://localhost:3000**. Health: http://localhost:3000/health.

---

## Environment

Create `.env` in this folder (e.g. from `env.example.sh`). Main variables:

- `PORT` — server port (default 3000)
- `HOST` — listen address (default 0.0.0.0)
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` — Postgres
- `REDIS_URL` — e.g. `redis://localhost:6379`
- `LOG_LEVEL` — `info` or `debug`
- `LOG_TO_FILE` — set to `0` to disable file logging

Example `.env`:

```bash
PORT=3000
HOST=0.0.0.0
NODE_ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_NAME=inventory
DB_USER=postgres
DB_PASSWORD=postgres
REDIS_URL=redis://localhost:6379
LOG_LEVEL=info
```

---

## Commands (this folder)

- `make dev` — run dev server (tsx watch)
- `make build` — production build to `dist/`
- `make start` — run `node dist/server.js`
- `make clean` — remove `dist/`
- `make check` — curl health endpoint
- `npm run lint` / `npm run lint:fix` — Biome

Infra (Docker, DB reset) is in parent: `make infra-up`, `make infra-down`, `make reset-db`, `make run-node`.

---

## API

Base: `http://localhost:3000/api/v1/inventory`

- `GET /stock/:sku` — current stock
- `POST /reserve` — naive (demo only)
- `POST /reserve/pessimistic` — SELECT FOR UPDATE
- `POST /reserve/optimistic` — version + retry
- `POST /reserve/redis` — Redis atomic counter

Reserve body:

```json
{ "sku": "SKU-TEST-001", "quantity": 1, "requestId": "<uuid>" }
```

---

## Reset DB and load tests

From **01-atomic-inventory**:

```bash
make reset-db
cd scripts/load-test
./race-test-naive.sh
./race-test-pessimistic.sh
./race-test-optimistic.sh
./race-test-redis.sh
```

See `../scripts/README.md` for how to interpret results.

---

## Logs

- Stdout: always (Pino).
- File: `logs/run-<ISO-timestamp>.log` when file logging is on (default unless `LOG_TO_FILE=0`).
- Disable file: `LOG_TO_FILE=0 npm run dev`

---

## implementation-go

Planned. See [implementation-go/README.md](../implementation-go/README.md).
