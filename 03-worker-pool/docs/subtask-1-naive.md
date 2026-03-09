# 1. Naive "Goroutine per Task"

**What:** A login (or generic) handler that spawns a new goroutine for every incoming task and returns `202 Accepted` immediately.

**Why:** To clearly demonstrate how the "just use goroutines" approach leads to unbounded resource usage, lack of guarantees, and problems under high RPS.

---

## Implementation steps

1. **Define a simple task type**
   - Minimal struct with task ID and payload (e.g. user ID, email).
   - No worker pool yet — just simulate some work with `time.Sleep` and logging.

2. **Naive handler**
   - HTTP handler like:
     - Parse request and build a `Task`.
     - Call `go processTask(ctx, task)` directly from the handler.
     - Immediately return `202 Accepted` (or `200 OK`).
   - `processTask`:
     - Simulate I/O (sleep, maybe fake DB call).
     - Print logs with task ID and goroutine ID.

3. **Demonstrate the problem**
   - Add a simple benchmark / load script (e.g. `wrk`, `hey`, or Go `testing.B`) that sends 10k–100k requests.
   - Observe:
     - Huge number of goroutines (`/debug/pprof/goroutine` or runtime metrics).
     - Growing memory usage.
     - Increased latencies and errors from downstream dependencies.

4. **Document limitations**
   - Note that there is:
     - No upper bound on parallel tasks.
     - No queueing/backpressure: everything is "fire and forget".
     - No way to safely shutdown while tasks are in flight.

---

## What will be done

- Naive handler with `go processTask(...)` per request.
- Simple task function that simulates some work.
- Load test or benchmark showing:
  - Exploding goroutine count.
  - Memory growth and instability.
- Clear explanation why this pattern is **not acceptable** for production.
