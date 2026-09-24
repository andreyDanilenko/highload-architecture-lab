# 2. Bounded Worker Pool

> Roadmap for extending the current implementation. The current pool uses a bounded channel and `Stop()` without a deadline; it cancels ffmpeg work but drains simulated work. The result queue, recovery and shutdown policy below are design targets, not completed features.

**What:** Replace "goroutine per task" with a fixed-size worker pool and a bounded in-memory queue (`chan *Task`), plus graceful shutdown.

**Why:** To bound parallelism and memory usage, and to define how a planned shutdown drains or cancels work. An in-memory queue cannot preserve jobs across a process crash.

---

## Implementation steps

1. **Introduce `WorkerPool`**
   - Implement a `WorkerPool` struct with:
     - Config: `numWorkers`, `queueSize`, `shutdownTimeout`.
     - Channels: `taskQueue chan *Task`, `resultQueue chan *Result`.
     - Lifecycle: `ctx`, `cancel`, `sync.WaitGroup`.
   - Provide `NewWorkerPool(ctx, numWorkers, queueSize, logger)` constructor.

2. **Start workers**
   - `Start()` method:
     - Spawn `numWorkers` goroutines running `wp.worker(id)`.
     - Optionally start a `resultHandler` and `monitor` goroutine.
   - `worker` loop:
     - Select on `wp.ctx.Done()` and `task := <-wp.taskQueue`.
     - For each task, call `processTaskWithRecover(id, task)`.

3. **Submit API and queue limits**
   - `Submit(task *Task) error`:
     - Non-blocking or bounded blocking enqueue into `taskQueue`.
     - On full queue, return `ErrQueueFull` (or similar) — do not allocate more memory.
   - HTTP handler:
     - Build `Task` from request.
     - Call `pool.Submit(task)`.
     - On `ErrQueueFull`, return `503 Service Unavailable` or `429` (backpressure).

4. **Graceful shutdown**
   - Implement `Shutdown(ctx context.Context) error`:
     - Stop accepting new work. For a drain policy, close admission and let accepted work finish before cancelling; immediate cancellation instead chooses to abandon unfinished work.
     - Wait for `wp.wg.Wait()` with timeout from ctx.
     - Log whether shutdown was graceful or timed out.
   - Wire it into `main()` with OS signal handling (SIGINT/SIGTERM).

---

## What will be done

- Introduce a `WorkerPool` abstraction with:
  - Fixed number of workers.
  - Bounded queue with backpressure via `ErrQueueFull`.
- Replace naive `go` usage in handler with `pool.Submit`.
- Add graceful shutdown so that:
  - In-flight tasks finish within the configured drain budget, or are explicitly reported as cancelled/unresolved.
  - No new tasks are accepted after shutdown starts.
