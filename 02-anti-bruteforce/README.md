# Task 02: Anti-Bruteforce Vault

Protect a critical endpoint (e.g. `/login`) from password bruteforce using a **sliding window** rate limit (Sliding Window Log), implemented in Redis with Lua for atomicity and performance.

---

## Problem

Simple "reset counter after N seconds" (TTL) is vulnerable: an attacker can send N requests in the last second of the window and N in the first second of the next, effectively doubling the limit. The sliding window counts only requests in the **last N seconds** at any moment.

---

## Task (overview)

Implement and compare four strategies:

1. **Naive** — in-memory store (Map + mutex). Demo: concurrency and scaling issues; limit is per instance.
2. **Pessimistic** — Redis distributed lock per IP; serialize read/update/release. No races; high latency and lock contention.
3. **Optimistic** — Redis WATCH/MULTI/EXEC; retry on conflict. No locks; many retries under contention (e.g. one IP hammering).
4. **Atomic** — single Lua script on Redis: trim old entries, count, add if under limit. One round-trip, fully atomic. **Production approach.**

Step-by-step plans per subtask are in `docs/`:

- [docs/redis-explained-step-by-step.md](docs/redis-explained-step-by-step.md) — **как работает Redis, L1/L2/L3 пошагово**
- [docs/subtask-1-naive.md](docs/subtask-1-naive.md) — in-memory, demo only
- [docs/subtask-2-pessimistic.md](docs/subtask-2-pessimistic.md) — Redis lock
- [docs/subtask-3-optimistic.md](docs/subtask-3-optimistic.md) — Redis WATCH + retry
- [docs/subtask-4-atomic.md](docs/subtask-4-atomic.md) — Redis Lua (sliding window)
- [docs/strategies-overview.md](docs/strategies-overview.md) — comparison of all four strategies

---

## Run

```bash
cd go && make dev
```

Optional: copy `.env.example` to `.env` in the project root. Redis is required only for strategies 2–4 (pessimistic, optimistic, atomic); the naive strategy runs without Redis.

---

## API (expected)

- `POST /login` or `POST /resource/naive` — naive (in-memory), demo only.
- `POST /resource/pessimistic` — lock-based.
- `POST /resource/optimistic` — WATCH-based.
- `POST /vault/login` — Lua-based sliding window (e.g. 5 attempts per IP per 60s).

Response: 200 — allowed; 429 — too many requests; 500 — server/Redis error (e.g. fail-close when Redis down).

---

## Testing

- **Naive:** 10 sequential requests from one IP → 6th gets 429. Two instances → effective limit doubled.
- **Pessimistic / Optimistic:** 100 concurrent requests from one IP → measure latency and retries.
- **Atomic:** 100 concurrent from one IP → 5× 200, 95× 429; 100 IPs → 100× 200; two instances → same limit shared.

---

## Problems and limitations per strategy

| Strategy    | Main problems / risks |
|-------------|------------------------|
| **Naive**   | Races and lock contention in process; does not scale (limit per instance). Demo only. |
| **Pessimistic** | High latency; lock contention; "herd blocking" when many requests wait for same lock. |
| **Optimistic**  | Many retries under contention (same IP); extra Redis load. |
| **Atomic**  | Lua debugging; Redis is single point of failure (fail-close or fail-open policy). |
