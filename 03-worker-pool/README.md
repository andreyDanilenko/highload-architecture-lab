# Task 03: Heavy Task Worker Pool

**Материалы для разбора:** [статьи EN/RU и реальные проекты](../docs/reading-map-01-30.md#task-03) · [оценка постановки](../docs/quality-review-01-30.md) · [связи с EventLab](../docs/project-playground-01-107.md).

Process background work through a **bounded worker pool**: cap parallelism, bound memory, and apply backpressure instead of spawning unbounded goroutines per request.

---

## Problem

`go processTask(...)` on every request looks cheap, but under a spike the number of goroutines and in-flight I/O grows without limit. Downstream systems (DB, Redis, HTTP clients) get hammered; memory and scheduling cost dominate; there is no queue, no overload signal, and no clean shutdown story. A worker pool turns fire-and-forget into a **managed** subsystem: fixed workers, bounded queue, and explicit behavior when the queue is full.

---

## Task (overview)

Compare the current implementations and the planned extensions below. All current queues are in memory: accepting a task does not make it durable across a process crash.

1. **Naive** — one goroutine per task from the handler (`202 Accepted` immediately). Demo: unbounded concurrency, lost control under load.
2. **Bounded — implemented:** fixed workers + bounded channel; `Dispatch` reports a full queue. `Stop` closes the queue and waits; ffmpeg work is cancelled, while simulated work drains. There is no shutdown deadline in the pool API.
3. **Reliable — partial:** bounded pool with panic recovery. Per-task timeouts, retry/backoff, Prometheus metrics and meaningful `/live`/`/ready` probes remain planned.
4. **Advanced — placeholder:** wraps the bounded pool with separate sizing. Priorities, circuit breaker, rate limiting and dynamic scaling remain planned.

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

## Current API

- `POST /work/naive` — enqueue via `go` from handler; demo only.
- `POST /work/bounded` — submit into bounded pool; backpressure when queue is full.
- `POST /work/reliable` — bounded admission and panic recovery; timeouts/retries/metrics are not implemented by this pool yet.
- `POST /work/advanced` — delegates to the bounded pool; no priorities or dynamic scaling yet.

Current responses: `202 Accepted` — accepted into in-memory execution, not confirmation of completion; `503 Service Unavailable` — queue full; `500 Internal Server Error` — dispatch error; `405 Method Not Allowed` — unsupported method.

---

## Verification targets

These are experiment plans, not a report of passed checks. Planned features require implementation before their assertions can pass.

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
| **Reliable** | Currently adds panic recovery, not durable execution or retries. Planned retries require idempotent effects and a retry budget. |
| **Advanced** | Currently shares bounded behavior. Planned prioritization and scaling add starvation, fairness and tuning risks. |
