# Worker Pool Strategies Overview

**Goal:** Compare bounded admission, failure handling and planned observability under load. Unbounded task creation can exhaust resources when arrivals outpace completion; a fixed worker count bounds task concurrency, but does not by itself guarantee durability or bound every resource.

**Status:** Bounded admission and reliable-pool panic recovery are implemented. Reliable timeouts/retries/metrics and Advanced priorities/scaling are roadmap designs below; `advanced.Pool` currently delegates to `bounded.Pool`. All current queues lose unfinished work on process crash.

---

## 1. Naive (Goroutine per Task)

- **Flow:** On each incoming request, the handler calls `go processTask(...)` directly. There is no queue, no limits, no central coordination.
- **Concurrency:** Active goroutines depend on arrival rate and task duration; task admission does not impose a concurrency limit.
- **Failure model:** If the process crashes, in-flight tasks are simply lost. Panics inside goroutines may kill the process if not recovered.
- **Scaling:** Measure saturation for the chosen workload and resources. There is no universal RPS threshold at which scheduling or a dependency must dominate.
- **Use:** Demo only — to demonstrate why "goroutine per request" is dangerous.

---

## 2. Bounded Worker Pool

- **Flow:** The handler does not start goroutines directly. It sends tasks into a **bounded channel** (queue), where a fixed number of workers pull and process them.
- **Concurrency:** Active tasks are capped by `numWorkers`, and queued task count by `queueSize`. Total memory still depends on task payloads, buffers and allocations inside the work.
- **Failure model:** Current `Stop()` closes the queue and waits without a deadline; it cancels ffmpeg work and drains simulated work. Panic recovery is currently in Reliable, not Bounded. There is no durable result queue.
- **Scaling:** At most `numWorkers` tasks run in this pool. Per-dependency concurrency can be higher if tasks fan out, and other pools/instances add their own load.
- **Use:** Baseline for bounded in-process execution. Jobs that must survive crashes require persistence, acknowledgment and recovery.

---

## 3. Reliable Worker Pool (Planned Timeouts, Retries, Metrics)

- **Flow:** Worker wraps execution in `context.WithTimeout`, tracks duration, and writes a `Result` into a result queue. Failed tasks can be retried using an exponential backoff strategy.
- **Reliability:** 
  - Timeouts signal cancellation; tasks must honor the context. They cannot forcibly stop arbitrary Go code.
  - `defer recover()` in workers prevents a single panic from killing the whole pool.
  - Retries may recover transient failures; cap attempts and total retry traffic, and require idempotent effects to avoid amplifying an outage.
- **Observability:** Prometheus metrics (`queueLength`, `activeWorkers`, `taskDuration`, `tasksProcessed`, `tasksFailed`, `taskRetries`) and health endpoints (`/live`, `/ready`) reveal pool state.
- **Use:** When you need **reliable** background processing with proper SLIs/SLOs.

---

## 4. Advanced Pool (Planned Priorities, Backpressure, Scaling)

- **Flow:** Tasks are enqueued into a `PriorityQueue` (multiple queues by priority). Workers always prefer higher-priority queues. Circuit breaker and rate limiter protect external services and the pool itself.
- **Backpressure:** 
  - Bounded queues + `ErrQueueFull` allow the system to reject new work under overload instead of crashing.
  - Circuit breaker (`CircuitBreaker.Execute`) short-circuits calls to broken downstream services.
- **Scaling:** A `DynamicPool` monitors metrics (queue load, active workers) and increases or decreases worker count between `minWorkers` and `maxWorkers`.
- **Use:** High-load / SRE-friendly production environments where prioritization, overload protection, and auto-scaling matter.

---

## Summary table

| Strategy    | Concurrency         | Reliability                        | Use Case                          | Main trade-off                    |
|------------|---------------------|------------------------------------|-----------------------------------|-----------------------------------|
| Naive      | Unbounded           | Tasks lost on crash; no limits     | Demo of what **not** to do       | Uncontrolled resource usage       |
| Bounded    | Fixed workers       | In-memory admission; Stop has no deadline | Baseline worker pool | Work can be lost on crash |
| Reliable   | Fixed workers       | Panic recovery now; timeouts/retries/metrics planned | Failure-handling experiments | No durable execution guarantee |
| Advanced   | Fixed workers now; dynamic/prioritized planned | Bounded wrapper now | Planned overload/fairness experiments | Higher complexity after implementation |
