# Worker Pool Strategies Overview

**Goal:** Process background tasks reliably under load using a **bounded worker pool**, with clear limits, observability, and overload protection. Why not just `go func()` per request? Because under spikes it explodes in goroutines, kills DB/Redis, and makes the system uncontrollable. The worker pool turns "fire and forget" into a managed, observable subsystem.

---

## 1. Naive (Goroutine per Task)

- **Flow:** On each incoming request, the handler calls `go processTask(...)` directly. There is no queue, no limits, no central coordination.
- **Concurrency:** Number of goroutines grows with RPS; nothing bounds parallelism.
- **Failure model:** If the process crashes, in-flight tasks are simply lost. Panics inside goroutines may kill the process if not recovered.
- **Scaling:** Under 10k–100k RPS, the runtime spends memory and CPU mostly on goroutine scheduling; DB/Redis are overloaded with uncontrolled parallelism.
- **Use:** Demo only — to demonstrate why "goroutine per request" is dangerous.

---

## 2. Bounded Worker Pool

- **Flow:** The handler does not start goroutines directly. It sends tasks into a **bounded channel** (queue), where a fixed number of workers pull and process them.
- **Concurrency:** Parallelism is capped by `numWorkers`. The queue has limited size (`queueSize`), so memory usage is bounded.
- **Failure model:** On shutdown, workers finish what they are doing (`Shutdown(ctx)` with timeout). Panics in workers are recovered and turned into `Result`.
- **Scaling:** Stable under load if limits are tuned; DB/Redis see at most `numWorkers` concurrent operations.
- **Use:** Baseline production pattern for any background processing.

---

## 3. Reliable Worker Pool (Timeouts, Retries, Metrics)

- **Flow:** Worker wraps execution in `context.WithTimeout`, tracks duration, and writes a `Result` into a result queue. Failed tasks can be retried using an exponential backoff strategy.
- **Reliability:** 
  - Timeouts prevent "stuck forever" tasks.
  - `defer recover()` in workers prevents a single panic from killing the whole pool.
  - Retries with backoff and jitter reduce transient error impact without hammering downstream services.
- **Observability:** Prometheus metrics (`queueLength`, `activeWorkers`, `taskDuration`, `tasksProcessed`, `tasksFailed`, `taskRetries`) and health endpoints (`/live`, `/ready`) reveal pool state.
- **Use:** When you need **reliable** background processing with proper SLIs/SLOs.

---

## 4. Advanced Pool (Priorities, Backpressure, Scaling)

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
| Bounded    | Fixed workers       | Graceful shutdown, panic recovery  | Baseline worker pool              | No retries / backoff yet          |
| Reliable   | Fixed workers       | Timeouts, retries, metrics, health | Production-ready reliability      | More code and configuration       |
| Advanced   | Dynamic, prioritized| Backpressure, CB, scaling, PQ      | High-load, business-critical load | Higher complexity, tuning needed  |
