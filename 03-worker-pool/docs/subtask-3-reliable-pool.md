# 3. Reliable Worker Pool

> Implementation status: the current reliable pool adds panic recovery to bounded execution. Per-task timeouts, retries, metrics and health probes below are planned work. Context cancellation is cooperative; in-memory acceptance does not survive a process crash.

**What:** Extend the bounded worker pool to be production-oriented: task timeouts, panic recovery, retries with exponential backoff + jitter, metrics, and health probes.

**Why:** Under real load, tasks can hang, downstream dependencies can be flaky, and failures must be observable. “Reliable” here means predictable failure behavior + observability, not “never fails”.

---

## Target behavior (acceptance criteria)

- **Bounded resources**: still only `workers` concurrent tasks; queue is bounded; overload returns `queue_full`.
- **Timeouts**: each task execution has a timeout; hung tasks do not block a worker forever.
- **Panic safety**: a panic inside a task does not kill the worker pool.
- **Retries**: transient failures are retried with backoff + jitter (limits to prevent retry storms).
- **Metrics**: Prometheus metrics expose queue depth, active workers, processed/failed/retried counters, duration histogram.
- **Health probes**:
  - **`/live`**: returns `200` if process is alive (no dependency on downstream systems).
  - **`/ready`**: returns `200` if the pool is ready to accept work (e.g. not shutting down, queue not critically full); otherwise `503`.

---

## Implementation plan

### 1) Add configuration knobs

- **Task timeout**: e.g. `WORKER_POOL_TASK_TIMEOUT_MS`.
- **Retry policy**: max attempts, initial/max backoff, jitter percent.
- **Readiness thresholds**: e.g. “ready while queue load < 90%”.

Where:
- `go/internal/config/config.go`

### 2) Define “work” result and error classification

- Decide what is retriable:
  - transient errors (network, 5xx, `context.DeadlineExceeded` from downstream calls) → retriable
  - validation/4xx/unsupported input → non-retriable
  - context canceled by shutdown → do not retry
- Provide helper: `isRetryable(err) bool`.

Where:
- `go/internal/work/` (or `go/internal/adapter/outbound/pool/reliable/` if you want it pool-local)

### 3) Wrap task execution with timeout + panic recovery

- In worker loop:
  - create `taskCtx, cancel := context.WithTimeout(poolCtx, taskTimeout)`
  - execute actual job with `taskCtx`
  - `defer recover()` around the job execution

Where:
- `go/internal/adapter/outbound/pool/reliable/pool.go`

### 4) Implement retries with exponential backoff + jitter

- Algorithm:
  - attempt 1 runs immediately
  - attempt \(n>1\): sleep(backoff(attempt-1) ± jitter), then retry
  - stop when success, non-retriable error, shutdown context done, or attempts exceeded
- Backoff example:
  - `delay = min(maxBackoff, initialBackoff * 2^(attempt-2))`
  - add jitter: `delay = delay * (1 ± jitterPct)`

Where:
- `go/internal/adapter/outbound/pool/reliable/pool.go`

### 5) Add Prometheus metrics endpoint + pool metrics

- Add `/metrics` HTTP endpoint (Prometheus client).
- Add pool metrics (suggested names):
  - `workerpool_queue_length` (gauge)
  - `workerpool_active_workers` (gauge)
  - `workerpool_tasks_processed_total{strategy,status}` (counter)
  - `workerpool_task_retries_total{strategy}` (counter)
  - `workerpool_task_duration_seconds{strategy}` (histogram)

Where:
- `go/internal/adapter/inbound/http/routes/*` and/or server wiring
- metrics package under `go/internal/` (recommended) or inside reliable pool (minimal)

### 6) Add `/live` and `/ready` endpoints

- `/live`: always `200` (unless you intentionally want to fail on catastrophic internal state).
- `/ready`: returns `503` when:
  - pool is stopping/stopped, or
  - queue load is above threshold (degraded overload), or
  - optionally: circuit breaker open (if you later add it)

Where:
- `go/internal/adapter/inbound/http/handlers/health.go` (add new handler or extend)
- `go/internal/adapter/inbound/http/routes/health_routes.go` (register routes)

### 7) Wire it into DI and routes

- Ensure `/work/reliable` actually uses the “reliable” implementation.
- Ensure stop/shutdown path cancels worker context and the pool stops cleanly.

Where:
- `go/internal/di/app.go`
- `go/internal/adapter/inbound/http/handlers/work.go`

---

## Test plan (manual)

- **Timeout**: configure a tiny timeout and a long simulated work → tasks should fail fast, workers keep moving.
- **Panic**: make one task panic → worker keeps running; metrics show failure.
- **Retries**: simulate flaky work (fail N-1 times) → task eventually succeeds; retry counter increments.
- **Readiness**: overload the queue → `/ready` becomes `503`; when load drops, returns `200`.
- **Prometheus**: hit `/metrics` and verify all expected metrics exist and change over time.

---

## What is currently missing in the repository (gap checklist)

- [ ] Task timeouts per job
- [ ] Retry with backoff + jitter
- [ ] Prometheus metrics + `/metrics` endpoint
- [ ] `/live` and `/ready` endpoints with meaningful readiness logic
