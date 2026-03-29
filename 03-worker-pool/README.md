# Task 03: Heavy Task Worker Pool

Process background work through a **bounded worker pool**: cap parallelism, bound memory, and apply backpressure instead of spawning unbounded goroutines per request.

---

## Problem

`go processTask(...)` on every request looks cheap, but under a spike the number of goroutines and in-flight I/O grows without limit. Downstream systems (DB, Redis, HTTP clients) get hammered; memory and scheduling cost dominate; there is no queue, no overload signal, and no clean shutdown story. A worker pool turns fire-and-forget into a **managed** subsystem: fixed workers, bounded queue, and explicit behavior when the queue is full.

---

## Task (overview)

Implement and compare four strategies:

1. **Naive** — one goroutine per task from the handler (`202 Accepted` immediately). Demo: unbounded concurrency, lost control under load.
2. **Bounded** — fixed workers + bounded channel; `Submit` returns error when the queue is full; graceful shutdown. Baseline production pattern.
3. **Reliable** — bounded pool plus task timeouts, panic recovery, retries with backoff/jitter, Prometheus metrics, `/live` and `/ready`. Production-oriented reliability.
4. **Advanced** — priorities, circuit breaker, rate limiting, dynamic worker count between min/max, explicit backpressure. High-load / SRE-style tuning.

Step-by-step plans per subtask are in `docs/`:

- [docs/subtask-1-naive.md](docs/subtask-1-naive.md) — goroutine-per-task demo
- [docs/subtask-2-bounded-pool.md](docs/subtask-2-bounded-pool.md) — bounded queue + workers
- [docs/subtask-3-reliable-pool.md](docs/subtask-3-reliable-pool.md) — timeouts, retries, metrics, health
- [docs/subtask-4-advanced-pool.md](docs/subtask-4-advanced-pool.md) — priorities, CB, scaling
- [docs/strategies-overview.md](docs/strategies-overview.md) — comparison of all four strategies
- [docs/spec/specification.md](docs/spec/specification.md) — implementation guide (best practices)
- [docs/spec/specification.ru.md](docs/spec/specification.ru.md) — same in Russian

---

## Run

From the task directory:

```bash
cp env.example .env   # optional
make dev
```

This runs `go run cmd/server/main.go` via the root `Makefile`. Optional `.env` is loaded from the project root (`go/internal/config` also tries `../.env` relative to `go/`).

---

## API (expected)

- `POST /work/naive` — enqueue via `go` from handler; demo only.
- `POST /work/bounded` — submit into bounded pool; backpressure when queue is full.
- `POST /work/reliable` — same as bounded path with timeouts/retries/metrics (pool implementation differs).
- `POST /work/advanced` — prioritized submission / advanced pool behavior.

Typical responses: `202 Accepted` / `200 OK` — task accepted or completed per design; `503 Service Unavailable` (or `429`) — queue full / overload; `500` — internal error.

---

## Testing

- **Naive:** high RPS load test → goroutine and memory growth; downstream saturation; contrast with bounded pool under the same load.
- **Bounded:** many concurrent submits → at most `numWorkers` tasks run in parallel; submits fail predictably when `queueSize` is exhausted; shutdown drains or times out per policy.
- **Reliable:** slow/failing work → retries and metrics move as expected; panics in workers do not kill the pool; health reflects degraded queue or error rate.
- **Advanced:** mixed priorities → higher-priority work preferred; circuit open → fast fail; scale-up/down reacts to queue load within min/max workers.

---

## Problems and limitations per strategy

| Strategy   | Main problems / risks |
|------------|------------------------|
| **Naive**  | Unbounded goroutines and load on dependencies; no backpressure; poor shutdown semantics. Demo only. |
| **Bounded** | Fixed capacity only; no retries/metrics until extended; tuning `numWorkers` / `queueSize` is workload-specific. |
| **Reliable** | More moving parts (backoff, metrics cardinality); retries can amplify load on a failing dependency if misconfigured. |
| **Advanced** | Highest complexity and tuning cost; priority inversion and CB/rate-limiter settings need careful design. |
