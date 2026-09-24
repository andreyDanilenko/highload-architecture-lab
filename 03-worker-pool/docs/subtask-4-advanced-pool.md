# 4. Advanced Worker Pool

**What:** Build a high-load oriented pool on top of the “reliable” baseline: priorities, explicit backpressure policies, rate limiting, circuit breaker, and dynamic worker scaling between min/max.

**Why:** When multiple task types compete under overload, you need control: protect critical work, shed non-critical load early, and avoid cascading failures on downstream systems.

---

## Target behavior (acceptance criteria)

- **Priority scheduling**: higher-priority tasks are processed first (no starvation or with explicit aging policy).
- **Explicit backpressure**: when queues are full → reject quickly and deterministically (HTTP `429` or `503`).
- **Rate limiting**: cap accepted task rate (per pool or per priority).
- **Circuit breaker**: when downstream is failing, short-circuit new work quickly (fail fast instead of piling up).
- **Dynamic scaling**: worker count scales between `minWorkers` and `maxWorkers` using queue load / utilization signal.

---

## Design choice (recommended minimal scope)

To keep the implementation tractable, start with **multiple bounded queues by priority** (e.g. `high`, `normal`, `low`) and a worker loop that always prefers higher priority:

- `queues[prio] chan Task`
- workers pick from highest prio first; if none, fall back to lower prio

This is simpler than a heap-based priority queue and easier to make bounded.

---

## Implementation plan

### 1) Define advanced task envelope with priority

- Either extend existing `domain.TaskID` dispatching by adding priority via:
  - query parameter (`/work/advanced?id=...&priority=high`), or
  - separate paths (`/work/advanced/high`, `/work/advanced/low`), or
  - a request header (`X-Task-Priority`)
- Normalize priority into a fixed set (e.g. 0..2).

Where:
- `go/internal/adapter/inbound/http/handlers/work.go` (parse priority)
- `go/internal/domain/` (if you introduce a new type)

### 2) Implement priority queues (bounded)

- Replace current `advanced.Pool` wrapper with a real implementation:
  - `queues []chan string` (one per priority)
  - configurable per-queue size (or split a shared capacity)
  - `Dispatch(ctx, taskID)` chooses queue by priority and enqueues non-blocking
- When the chosen queue is full: return `QueueFull` (explicit overload).

Where:
- `go/internal/adapter/outbound/pool/advanced/pool.go`

### 3) Worker selection logic (priority preference + fairness)

Baseline logic:
- try `high` queue non-blocking
- then `normal`
- then `low`
- if all empty: block on “any task available” (or short sleep + retry loop)

Add fairness to avoid starvation (pick one):
- **aging**: after K high tasks, force 1 normal/low if present
- or **weighted round-robin**: e.g. 5:3:1 picks

Where:
- `go/internal/adapter/outbound/pool/advanced/pool.go`

### 4) Rate limiting on acceptance

- Implement a token bucket limiter (or use `golang.org/x/time/rate`).
- Apply at `Dispatch` time (cheapest): if no token → reject fast (prefer HTTP `429`).

Where:
- `go/internal/adapter/outbound/pool/advanced/` (limiter)
- config: `WORKER_POOL_ADVANCED_RPS`, `WORKER_POOL_ADVANCED_BURST`

### 5) Circuit breaker for downstream calls

- Minimal version: CB wraps the actual work execution; if CB is open → fail fast (do not spend worker time).
- Track failures/successes; open after threshold; half-open after timeout.

Where:
- `go/internal/adapter/outbound/pool/advanced/` (cb implementation)

### 6) Dynamic worker scaling (min/max)

- Keep a baseline worker count `minWorkers`.
- Periodically compute queue load:
  - `load = totalQueued / totalCapacity`
- Scale up when load > `scaleUpThreshold`, scale down when load < `scaleDownThreshold` for some time.
- Scaling down needs a safe stop mechanism for extra workers (e.g. per-worker stop channels).

Where:
- `go/internal/adapter/outbound/pool/advanced/` (scaler loop + worker lifecycle)
- config: `ADV_MIN_WORKERS`, `ADV_MAX_WORKERS`, thresholds, interval

### 7) Metrics + readiness integration

- Export additional metrics (per-priority queue lengths, limiter drops, CB state).
- Ensure `/ready` accounts for overload (queue almost full, limiter constantly rejecting, CB open for too long).

Where:
- same metrics/health plumbing as in subtask 3

---

## Test plan (manual)

- **Priority**: flood low priority + trickle high → high tasks complete quickly even under load.
- **Fairness**: keep high constant + low constant → low still gets some throughput per chosen policy.
- **Rate limiting**: set low RPS → API starts rejecting quickly even if queue is empty.
- **Circuit breaker**: simulate downstream failing → pool fails fast; queue does not grow unbounded.
- **Scaling**: burst traffic → workers scale up to max; idle → scale down to min.

---

## What is currently missing in the repository (gap checklist)

- [ ] Real advanced pool implementation (currently just wraps `bounded`)
- [ ] Priority-aware dispatch + scheduling
- [ ] Rate limiting
- [ ] Circuit breaker
- [ ] Dynamic scaling (min/max workers)
- [ ] Metrics + readiness rules for advanced behavior
