# Best Practices for Worker Pool Implementation
## Universal Guide for Background Task Processing

---

## 1. Fundamental Problems and Their Solutions

### 1.1. Why Simple "Goroutine per Task" Kills the System

**Problem:** In Go, `go func()` seems cheap, but 100k goroutines will consume all memory. In Node.js, each async call isn't free either.

```go
// BAD - never do this in production
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
    // At 10k RPS, this creates 10k goroutines -> OOM killer
    go s.processTask(r.Context(), extractTask(r))
    w.WriteHeader(http.StatusAccepted)
}
```

**Why this is bad:**
- No concurrency limits
- DB/Redis will crash under load spikes
- No control over the queue
- Tasks are lost when server crashes

**Correct approach:** Worker pool with limited workers and managed queue.

---

## 2. Production-Ready Worker Pool Architecture

### 2.1. Component Structure

```mermaid
graph TD
    A[HTTP Handler] --> B[Task Dispatcher]
    B --> C[(Task Queue)]
    C --> D[Worker 1]
    C --> E[Worker 2]
    C --> F[Worker N]
    
    D --> G[(Result Queue)]
    E --> G
    F --> G
    
    G --> H[Result Handler]
    H --> I[Client/Webhook]
    
    J[Health Check] --> D
    J --> E
    J --> F
    
    K[Circuit Breaker] --> D
    K --> E
    K --> F
```

### 2.2. Basic Implementation with Full Control

```go
package workerpool

import (
    "context"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "go.uber.org/zap"
)

// Task defines a unit of work
type Task struct {
    ID          string
    Type        string
    Payload     interface{}
    Priority    int // 0-9, higher is more important
    CreatedAt   time.Time
    MaxRetries  int
    Timeout     time.Duration
}

// Result represents task execution outcome
type Result struct {
    TaskID    string
    Success   bool
    Error     error
    Duration  time.Duration
    Attempt   int
}

// WorkerPool main structure
type WorkerPool struct {
    // Configuration
    numWorkers   int
    queueSize    int
    
    // Communication channels
    taskQueue    chan *Task
    resultQueue  chan *Result
    
    // Lifecycle management
    ctx          context.Context
    cancel       context.CancelFunc
    wg           sync.WaitGroup
    
    // State
    activeWorkers int32
    tasksProcessed int64
    tasksFailed    int64
    
    // Dependencies
    logger       *zap.Logger
    metrics      *Metrics
    
    // Overload protection
    circuitBreaker *CircuitBreaker
    rateLimiter    *RateLimiter
    
    // Graceful shutdown
    shutdownTimeout time.Duration
}

// Metrics for Prometheus
type Metrics struct {
    queueLength     prometheus.Gauge
    activeWorkers   prometheus.Gauge
    taskDuration    *prometheus.HistogramVec
    tasksProcessed  *prometheus.CounterVec
    tasksFailed     *prometheus.CounterVec
    taskRetries     *prometheus.CounterVec
}

func NewWorkerPool(ctx context.Context, numWorkers, queueSize int, logger *zap.Logger) *WorkerPool {
    ctx, cancel := context.WithCancel(ctx)
    
    wp := &WorkerPool{
        numWorkers:      numWorkers,
        queueSize:       queueSize,
        taskQueue:       make(chan *Task, queueSize),
        resultQueue:     make(chan *Result, queueSize),
        ctx:             ctx,
        cancel:          cancel,
        logger:          logger,
        shutdownTimeout: 30 * time.Second,
    }
    
    wp.initMetrics()
    return wp
}

// Start launches workers
func (wp *WorkerPool) Start() {
    wp.logger.Info("starting worker pool", 
        zap.Int("workers", wp.numWorkers),
        zap.Int("queue_size", wp.queueSize))
    
    // Start workers
    for i := 0; i < wp.numWorkers; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
    
    // Start result handler
    wp.wg.Add(1)
    go wp.resultHandler()
    
    // Start monitoring
    wp.wg.Add(1)
    go wp.monitor()
}

// worker main processing loop
func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()
    atomic.AddInt32(&wp.activeWorkers, 1)
    defer atomic.AddInt32(&wp.activeWorkers, -1)
    
    wp.logger.Debug("worker started", zap.Int("worker_id", id))
    
    for {
        select {
        case <-wp.ctx.Done():
            wp.logger.Debug("worker stopping", zap.Int("worker_id", id))
            return
            
        case task := <-wp.taskQueue:
            wp.processTaskWithRecover(id, task)
        }
    }
}

// processTaskWithRecover with panic protection
func (wp *WorkerPool) processTaskWithRecover(workerID int, task *Task) {
    defer func() {
        if r := recover(); r != nil {
            wp.logger.Error("worker panic recovered",
                zap.Int("worker_id", workerID),
                zap.String("task_id", task.ID),
                zap.Any("panic", r),
                zap.Stack("stack"))
            
            // Send error result
            wp.resultQueue <- &Result{
                TaskID:  task.ID,
                Success: false,
                Error:   fmt.Errorf("panic: %v", r),
            }
        }
    }()
    
    start := time.Now()
    
    // Create context with timeout for task
    taskCtx, cancel := context.WithTimeout(wp.ctx, task.Timeout)
    defer cancel()
    
    // Process task
    err := wp.executeTask(taskCtx, task)
    
    duration := time.Since(start)
    
    // Send result
    result := &Result{
        TaskID:   task.ID,
        Success:  err == nil,
        Error:    err,
        Duration: duration,
    }
    
    // Don't block on result sending
    select {
    case wp.resultQueue <- result:
    default:
        wp.logger.Warn("result queue full, dropping result", 
            zap.String("task_id", task.ID))
    }
    
    // Update metrics
    wp.metrics.taskDuration.WithLabelValues(task.Type).Observe(duration.Seconds())
    atomic.AddInt64(&wp.tasksProcessed, 1)
    if err != nil {
        atomic.AddInt64(&wp.tasksFailed, 1)
    }
}
```

---

## 3. Critical Patterns for Production

### 3.1. Graceful Shutdown

```go
// Shutdown stops the pool with active task completion
func (wp *WorkerPool) Shutdown(ctx context.Context) error {
    wp.logger.Info("shutting down worker pool")
    
    // Signal stop
    wp.cancel()
    
    // Completion signal channel
    done := make(chan struct{})
    
    go func() {
        // Wait for all workers to finish
        wp.wg.Wait()
        close(done)
    }()
    
    // Wait either for completion or timeout
    select {
    case <-done:
        wp.logger.Info("worker pool stopped gracefully")
        return nil
    case <-ctx.Done():
        wp.logger.Warn("worker pool shutdown timeout")
        return ctx.Err()
    }
}

// Usage in main
func main() {
    wp := workerpool.NewWorkerPool(context.Background(), 10, 100, logger)
    wp.Start()
    
    // Wait for OS signals
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := wp.Shutdown(ctx); err != nil {
        log.Fatal("shutdown failed:", err)
    }
}
```

### 3.2. Task Prioritization

```go
// PriorityQueue - priority-based queue
type PriorityQueue struct {
    queues   []chan *Task
    mu       sync.RWMutex
    stopped  bool
}

func NewPriorityQueue(levels int, size int) *PriorityQueue {
    pq := &PriorityQueue{
        queues: make([]chan *Task, levels),
    }
    for i := 0; i < levels; i++ {
        pq.queues[i] = make(chan *Task, size)
    }
    return pq
}

// AddTask adds task with priority consideration
func (pq *PriorityQueue) AddTask(task *Task) error {
    pq.mu.RLock()
    defer pq.mu.RUnlock()
    
    if pq.stopped {
        return ErrQueueStopped
    }
    
    // Normalize priority to [0, levels-1] range
    priority := task.Priority
    if priority < 0 {
        priority = 0
    }
    if priority >= len(pq.queues) {
        priority = len(pq.queues) - 1
    }
    
    select {
    case pq.queues[priority] <- task:
        return nil
    default:
        return ErrQueueFull
    }
}

// GetTask retrieves highest priority task
func (pq *PriorityQueue) GetTask(ctx context.Context) (*Task, error) {
    for {
        for i := len(pq.queues) - 1; i >= 0; i-- {
            select {
            case task := <-pq.queues[i]:
                return task, nil
            default:
                continue
            }
        }
        
        // If nothing available, wait for any task
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
            time.Sleep(10 * time.Millisecond)
        }
    }
}
```

### 3.3. Retry with Exponential Backoff

```go
type RetryStrategy struct {
    MaxAttempts     int
    InitialInterval time.Duration
    MaxInterval     time.Duration
    Multiplier      float64
    Jitter          float64
}

func DefaultRetryStrategy() *RetryStrategy {
    return &RetryStrategy{
        MaxAttempts:     3,
        InitialInterval: 100 * time.Millisecond,
        MaxInterval:     10 * time.Second,
        Multiplier:      2.0,
        Jitter:          0.1, // 10% jitter
    }
}

func (wp *WorkerPool) executeWithRetry(task *Task) error {
    strategy := DefaultRetryStrategy()
    var lastErr error
    
    for attempt := 0; attempt < strategy.MaxAttempts; attempt++ {
        // Check context
        select {
        case <-wp.ctx.Done():
            return wp.ctx.Err()
        default:
        }
        
        // Execute task
        err := wp.executeTask(wp.ctx, task)
        if err == nil {
            return nil
        }
        
        lastErr = err
        wp.logger.Warn("task failed, will retry",
            zap.String("task_id", task.ID),
            zap.Int("attempt", attempt+1),
            zap.Error(err))
        
        // Count retry in metrics
        wp.metrics.taskRetries.WithLabelValues(task.Type).Inc()
        
        if attempt < strategy.MaxAttempts-1 {
            // Calculate delay with jitter
            delay := strategy.InitialInterval * time.Duration(
                math.Pow(strategy.Multiplier, float64(attempt)))
            if delay > strategy.MaxInterval {
                delay = strategy.MaxInterval
            }
            
            // Add jitter (±10%)
            jitter := time.Duration(float64(delay) * strategy.Jitter * 
                (rand.Float64()*2 - 1))
            delay += jitter
            
            time.Sleep(delay)
        }
    }
    
    return fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

---

## 4. Monitoring and Observability

### 4.1. Health Checks

```go
type WorkerPoolHealth struct {
    Status           string        `json:"status"`
    ActiveWorkers    int32         `json:"active_workers"`
    QueueLength      int           `json:"queue_length"`
    QueueCapacity    int           `json:"queue_capacity"`
    TasksProcessed   int64         `json:"tasks_processed"`
    TasksFailed      int64         `json:"tasks_failed"`
    ErrorRate        float64       `json:"error_rate"`
    AvgProcessingTime time.Duration `json:"avg_processing_time"`
}

func (wp *WorkerPool) HealthCheck() *WorkerPoolHealth {
    health := &WorkerPoolHealth{
        ActiveWorkers:  atomic.LoadInt32(&wp.activeWorkers),
        QueueLength:    len(wp.taskQueue),
        QueueCapacity:  cap(wp.taskQueue),
        TasksProcessed: atomic.LoadInt64(&wp.tasksProcessed),
        TasksFailed:    atomic.LoadInt64(&wp.tasksFailed),
    }
    
    // Calculate error rate
    if health.TasksProcessed > 0 {
        health.ErrorRate = float64(health.TasksFailed) / 
            float64(health.TasksProcessed)
    }
    
    // Determine status
    switch {
    case health.QueueLength > health.QueueCapacity*9/10:
        health.Status = "degraded" // queue nearly full
    case health.ErrorRate > 0.1:
        health.Status = "degraded" // >10% errors
    case health.ActiveWorkers == 0:
        health.Status = "down"
    default:
        health.Status = "healthy"
    }
    
    return health
}

// HTTP handler for k8s probes
func (wp *WorkerPool) LiveHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

func (wp *WorkerPool) ReadyHandler(w http.ResponseWriter, r *http.Request) {
    health := wp.HealthCheck()
    
    if health.Status == "healthy" {
        w.WriteHeader(http.StatusOK)
    } else {
        w.WriteHeader(http.StatusServiceUnavailable)
    }
    
    json.NewEncoder(w).Encode(health)
}
```

### 4.2. Detailed Metrics

```go
func (wp *WorkerPool) initMetrics() {
    wp.metrics = &Metrics{
        queueLength: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "workerpool_queue_length",
            Help: "Current queue length",
        }),
        
        activeWorkers: promauto.NewGauge(prometheus.GaugeOpts{
            Name: "workerpool_active_workers",
            Help: "Number of active workers",
        }),
        
        taskDuration: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "workerpool_task_duration_seconds",
                Help:    "Task processing duration",
                Buckets: prometheus.DefBuckets,
            },
            []string{"task_type"},
        ),
        
        tasksProcessed: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "workerpool_tasks_processed_total",
                Help: "Total tasks processed",
            },
            []string{"task_type", "status"},
        ),
        
        taskRetries: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "workerpool_task_retries_total",
                Help: "Total task retries",
            },
            []string{"task_type"},
        ),
    }
    
    // Regular gauge updates
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-wp.ctx.Done():
                return
            case <-ticker.C:
                wp.metrics.queueLength.Set(float64(len(wp.taskQueue)))
                wp.metrics.activeWorkers.Set(float64(
                    atomic.LoadInt32(&wp.activeWorkers)))
            }
        }
    }()
}
```

### 4.3. Structured Logging

```go
type TaskLog struct {
    Level       string    `json:"level"`
    Timestamp   time.Time `json:"@timestamp"`
    
    TaskID      string    `json:"task_id"`
    TaskType    string    `json:"task_type"`
    Priority    int       `json:"priority"`
    
    Attempt     int       `json:"attempt"`
    Duration    string    `json:"duration_ms"`
    
    WorkerID    int       `json:"worker_id"`
    QueueLength int       `json:"queue_length"`
    
    Error       string    `json:"error,omitempty"`
    StackTrace  string    `json:"stacktrace,omitempty"`
}

func (wp *WorkerPool) logTaskCompletion(task *Task, result *Result, workerID int) {
    logEntry := map[string]interface{}{
        "level":       "info",
        "@timestamp": time.Now(),
        "task_id":     task.ID,
        "task_type":   task.Type,
        "priority":    task.Priority,
        "duration_ms": result.Duration.Milliseconds(),
        "worker_id":   workerID,
        "queue_length": len(wp.taskQueue),
        "success":     result.Success,
    }
    
    if result.Error != nil {
        logEntry["level"] = "error"
        logEntry["error"] = result.Error.Error()
        
        // Log stack trace for unexpected errors
        if errors.Is(result.Error, context.DeadlineExceeded) {
            logEntry["stacktrace"] = string(debug.Stack())
        }
    }
    
    // JSON logging for ELK
    logJSON, _ := json.Marshal(logEntry)
    fmt.Println(string(logJSON))
}
```

---

## 5. Overload Protection (Backpressure)

### 5.1. Circuit Breaker for External Services

```go
type CircuitBreaker struct {
    mu              sync.RWMutex
    state           string // closed, open, half-open
    failures        int
    threshold       int
    timeout         time.Duration
    lastFailureTime time.Time
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
    // Check state
    cb.mu.RLock()
    state := cb.state
    cb.mu.RUnlock()
    
    switch state {
    case "open":
        // Check if timeout has passed
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.mu.Lock()
            cb.state = "half-open"
            cb.mu.Unlock()
        } else {
            return ErrCircuitOpen
        }
    }
    
    // Try to execute
    err := fn()
    
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    if err != nil {
        cb.failures++
        cb.lastFailureTime = time.Now()
        
        if cb.failures >= cb.threshold {
            cb.state = "open"
        }
        return err
    }
    
    // Success
    if cb.state == "half-open" {
        cb.state = "closed"
        cb.failures = 0
    }
    
    return nil
}
```

### 5.2. Dynamic Scaling (if needed)

```go
type DynamicPool struct {
    *WorkerPool
    minWorkers     int
    maxWorkers     int
    scaleUpFactor  float64 // when queue fills to N%
    scaleDownFactor float64 // when workers are idle
    monitorInterval time.Duration
}

func (dp *DynamicPool) monitorAndScale() {
    ticker := time.NewTicker(dp.monitorInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-dp.ctx.Done():
            return
        case <-ticker.C:
            queueLoad := float64(len(dp.taskQueue)) / float64(cap(dp.taskQueue))
            activeWorkers := int(atomic.LoadInt32(&dp.activeWorkers))
            
            switch {
            case queueLoad > dp.scaleUpFactor && activeWorkers < dp.maxWorkers:
                // Need more workers
                newWorkers := min(activeWorkers*2, dp.maxWorkers)
                dp.scaleTo(newWorkers)
                
            case queueLoad < 0.1 && activeWorkers > dp.minWorkers:
                // Too many workers idle
                newWorkers := max(activeWorkers/2, dp.minWorkers)
                dp.scaleTo(newWorkers)
            }
        }
    }
}
```

---

## 6. Configuration for Different Scenarios

### 6.1. Production Configuration (YAML)

```yaml
worker_pool:
  # Basic parameters
  num_workers: 10          # Start with 10, adjust based on metrics
  queue_size: 1000         # Maximum tasks in queue
  
  # Processing strategy
  processing_strategy: "parallel"  # parallel or sequential
  
  # Queue full behavior
  queue_full_policy: "reject"  # reject, wait, or drop_oldest
  
  # Prioritization
  priority_levels: 5
  priority_default: 2
  
  # Timeouts
  task_timeout: "30s"
  shutdown_timeout: "30s"
  
  # Retry policy
  retry:
    max_attempts: 3
    initial_interval: "100ms"
    max_interval: "10s"
    multiplier: 2.0
    jitter: 0.1
  
  # Task rate limiting
  rate_limit:
    enabled: true
    tasks_per_second: 100
    burst: 20
  
  # Circuit breaker for external calls
  circuit_breaker:
    enabled: true
    failure_threshold: 5
    timeout: "10s"
  
  # Monitoring
  health_check_interval: "10s"
  metrics_enabled: true
  tracing_enabled: true
  
  # Logging
  log_level: "info"
  log_slow_tasks_threshold: "5s"
```

### 6.2. Feature Flags for A/B Testing

```go
type WorkerPoolConfig struct {
    // New feature enablement
    EnablePriorityQueue   bool `feature:"priority-queue"`
    EnableDynamicScaling  bool `feature:"dynamic-scaling"`
    EnableCircuitBreaker  bool `feature:"circuit-breaker"`
    
    // Gradual rollout
    CanaryPercent        int    `feature:"new-worker-pool"`
    CanaryTaskTypes      []string // Only for specific task types
    
    // Dynamic parameters
    WorkersByHour        map[int]int // Different worker counts by hour
    QueueSizeByLoad      map[string]int // Queue size based on load
}
```

---

## 7. Testing

### 7.1. Load Testing

```go
func TestWorkerPoolUnderLoad(t *testing.T) {
    pool := NewWorkerPool(context.Background(), 10, 100, logger)
    pool.Start()
    
    // Generate load
    const numTasks = 10000
    var wg sync.WaitGroup
    
    start := time.Now()
    
    for i := 0; i < numTasks; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            task := &Task{
                ID:        fmt.Sprintf("task-%d", id),
                Type:      "test",
                Payload:   id,
                CreatedAt: time.Now(),
            }
            
            err := pool.Submit(task)
            if err != nil {
                t.Logf("submit error: %v", err)
            }
        }(i)
    }
    
    wg.Wait()
    
    // Wait for processing completion
    time.Sleep(5 * time.Second)
    
    duration := time.Since(start)
    t.Logf("Processed %d tasks in %v", numTasks, duration)
    t.Logf("Throughput: %.2f tasks/sec", float64(numTasks)/duration.Seconds())
    
    // Check metrics
    health := pool.HealthCheck()
    if health.TasksFailed > 0 {
        t.Errorf("%d tasks failed", health.TasksFailed)
    }
}
```

### 7.2. Edge Case Testing

```go
func TestWorkerPoolEdgeCases(t *testing.T) {
    t.Run("queue full rejection", func(t *testing.T) {
        pool := NewWorkerPool(context.Background(), 1, 5, logger)
        pool.Start()
        
        // Fill queue
        for i := 0; i < 10; i++ {
            err := pool.Submit(&Task{ID: fmt.Sprintf("task-%d", i)})
            if i < 5 && err != nil {
                t.Errorf("expected nil, got %v", err)
            }
            if i >= 5 && err != ErrQueueFull {
                t.Errorf("expected ErrQueueFull, got %v", err)
            }
        }
    })
    
    t.Run("worker panic recovery", func(t *testing.T) {
        pool := NewWorkerPool(context.Background(), 1, 5, logger)
        pool.Start()
        
        // Task that panics
        err := pool.Submit(&Task{
            ID: "panic-task",
            Payload: func() {
                panic("test panic")
            },
        })
        
        if err != nil {
            t.Fatal(err)
        }
        
        // Wait for processing
        time.Sleep(100 * time.Millisecond)
        
        // Check that worker is still alive
        health := pool.HealthCheck()
        if health.ActiveWorkers != 1 {
            t.Errorf("worker died after panic")
        }
    })
}
```

---

## 8. Production Checklist

### ✅ Must have (critical)
- [ ] Limit on parallel tasks (worker pool)
- [ ] Graceful shutdown with active task completion
- [ ] Panic protection in workers (recover)
- [ ] Task execution timeouts
- [ ] Bounded queue size
- [ ] Prometheus metrics (active workers, queue length, execution time)
- [ ] Health checks for k8s
- [ ] Error logging

### ⚠️ Should have (for reliability)
- [ ] Retry with exponential backoff
- [ ] Circuit breaker for external services
- [ ] Rate limiting for task submission
- [ ] Task prioritization
- [ ] Metrics by task type
- [ ] Slow task monitoring
- [ ] Graceful degradation on queue overflow

### 🚀 Nice to have (for high loads)
- [ ] Dynamic worker scaling
- [ ] Persistent queue (Redis/Kafka) for durability
- [ ] Distributed worker pool across instances
- [ ] Task pool isolation (CPU-bound vs I/O-bound)
- [ ] Request tracing through worker pool
- [ ] Auto-tuning based on load

---

## Conclusion

A worker pool isn't just a "goroutine pool" but a critical production system component. Key principles:

1. **Always bound resources** — never let tasks consume unlimited resources
2. **Fail gracefully** — reject under overload, don't crash
3. **Observe everything** — without metrics you're blind
4. **Test failure modes** — task panic shouldn't kill the worker
5. **Shutdown gracefully** — can't lose tasks in production

