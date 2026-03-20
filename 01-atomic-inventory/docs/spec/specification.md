# Best Practices for Atomic Inventory Counter
## A Comprehensive Guide to Concurrent Resource Access

---

## 1. Understanding the Problem and Business Requirements

### 1.1. Real-world Use Cases
In production, this could be:
- **Airline tickets**: 200 seats, 5000 people trying to book simultaneously
- **Concert tickets**: Last-minute discounted tickets for the first 100 customers
- **E-commerce**: iPhone flash sale "limited stock"
- **Hotel bookings**: One room, two simultaneous reservation attempts

### 1.2. Core Requirements
```go
type InventoryRequirements struct {
    // Hard requirements (Must have)
    NoNegative      bool  // Never go negative
    AtomicDecrement bool  // Two parallel requests must not take 1 item twice
    Consistency     bool  // After successful decrement, item is truly reserved
    
    // Business requirements
    Overbooking     bool  // Can we sell more than available? (usually NO)
    WaitingList     bool  // Queue for out-of-stock items
    PartialAllowed  bool  // Can buy 3 if only 2 left? (usually NO)
}
```

---

## 2. Concurrency Control Strategies

### 2.1. Strategy Comparison

| Strategy | Mechanism | Pros | Cons | When to Use |
|----------|-----------|------|------|-------------|
| **Pessimistic Lock** | `SELECT FOR UPDATE` | Guaranteed integrity, simple to understand | Locks rows, can deadlock, poor under high load | Medium load, consistency critical |
| **Optimistic Lock** | Version column + `UPDATE WHERE version = ?` | No locks, scales well | Requires retry logic, many conflicts possible | High load, low contention |
| **Atomic DB Ops** | `UPDATE SET count = count - 1 WHERE count > 0` | Maximum performance | Limited logic, hard with additional actions | Simple counters, no extra logic needed |
| **Redis** | `DECR` + watch | Microsecond latency | Potential data loss, no DB transactions | Caching counters, temporary sales |
| **Queue-based** | Order queue + worker | Full control, no race conditions | Latency, complexity | Async responses acceptable |

### 2.2. Production-ready PostgreSQL Implementation

```sql
-- Inventory table
CREATE TABLE inventory (
    id BIGSERIAL PRIMARY KEY,
    sku VARCHAR(50) NOT NULL,
    quantity INT NOT NULL CHECK (quantity >= 0),
    reserved INT NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Database-level constraints (last line of defense)
    CONSTRAINT quantity_non_negative CHECK (quantity >= 0),
    CONSTRAINT reserved_non_negative CHECK (reserved >= 0),
    CONSTRAINT available_check CHECK (quantity - reserved >= 0)
);

CREATE INDEX idx_inventory_sku ON inventory(sku);

-- Function for atomic decrement with validation
CREATE OR REPLACE FUNCTION atomic_decrement(
    p_sku VARCHAR,
    p_quantity INT
) RETURNS TABLE (
    success BOOLEAN,
    new_quantity INT,
    old_version BIGINT
) LANGUAGE plpgsql AS $$
DECLARE
    v_current inventory%ROWTYPE;
BEGIN
    -- Pessimistic lock the row
    SELECT * INTO v_current 
    FROM inventory 
    WHERE sku = p_sku 
    FOR UPDATE;
    
    IF NOT FOUND THEN
        RETURN QUERY SELECT false, 0, 0::BIGINT;
        RETURN;
    END IF;
    
    -- Check if enough stock
    IF v_current.quantity - v_current.reserved < p_quantity THEN
        RETURN QUERY SELECT false, v_current.quantity, v_current.version;
        RETURN;
    END IF;
    
    -- Update with version check (optimistic lock as extra protection)
    UPDATE inventory 
    SET 
        quantity = quantity - p_quantity,
        version = version + 1,
        updated_at = NOW()
    WHERE 
        sku = p_sku 
        AND version = v_current.version  -- Optimistic lock as backup
        AND quantity >= p_quantity;      -- SQL-level protection
    
    GET DIAGNOSTICS v_current = ROW_COUNT;
    
    IF v_current > 0 THEN
        RETURN QUERY SELECT true, quantity, version + 1 
        FROM inventory WHERE sku = p_sku;
    ELSE
        RETURN QUERY SELECT false, quantity, version 
        FROM inventory WHERE sku = p_sku;
    END IF;
END;
$$;
```

---

## 3. Multi-layer Protection Against Race Conditions

### 3.1. Defense Layers Diagram

```mermaid
graph TD
    A[Client Request] --> B[Application Lock Layer]
    
    subgraph "Layer 1: Distributed Lock"
        B1[Redis Redlock] --> B
    end
    
    subgraph "Layer 2: Database Transaction"
        B2[BEGIN TRANSACTION<br/>ISOLATION LEVEL REPEATABLE READ] --> B
    end
    
    subgraph "Layer 3: Row-Level Lock"
        B3[SELECT FOR UPDATE] --> B
    end
    
    subgraph "Layer 4: Optimistic Lock"
        B4[WHERE version = ?] --> B
    end
    
    subgraph "Layer 5: Constraint"
        B5[CHECK quantity >= 0] --> B
    end
    
    B --> C[Success/Failure]
```

### 3.2. Go Implementation with Multi-layer Protection

```go
type InventoryService struct {
    db        *sql.DB
    redis     *redis.Client
    metrics   *InventoryMetrics
    logger    *slog.Logger
}

type DecrementRequest struct {
    SKU       string
    Quantity  int
    RequestID string // for idempotency
    UserID    string // for audit
}

type DecrementResult struct {
    Success     bool
    NewQuantity int
    OldVersion  int64
    Error       error
    Retryable   bool
}

func (s *InventoryService) AtomicDecrement(ctx context.Context, req *DecrementRequest) (*DecrementResult, error) {
    startTime := time.Now()
    defer func() {
        s.metrics.ObserveLatency(startTime)
    }()

    // Layer 0: Input validation
    if req.Quantity <= 0 {
        return nil, ErrInvalidQuantity
    }

    // Layer 1: Distributed lock to prevent double spending
    lockKey := fmt.Sprintf("inventory:lock:%s", req.SKU)
    lock, err := s.acquireLock(ctx, lockKey, 5*time.Second)
    if err != nil {
        // If lock fails, try without it - not critical
        s.logger.Warn("failed to acquire distributed lock", "sku", req.SKU, "error", err)
    } else {
        defer lock.Release()
    }

    // Layer 2: Database transaction with proper isolation level
    tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
        Isolation: sql.LevelRepeatableRead,
        ReadOnly:  false,
    })
    if err != nil {
        return nil, fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback() // safe rollback if commit not called

    // Layer 3: Pessimistic lock + versioning
    var current struct {
        Quantity int
        Reserved int
        Version  int64
    }
    
    err = tx.QueryRowContext(ctx, `
        SELECT quantity, reserved, version 
        FROM inventory 
        WHERE sku = $1 
        FOR UPDATE  -- pessimistic lock
    `, req.SKU).Scan(&current.Quantity, &current.Reserved, &current.Version)
    
    if err == sql.ErrNoRows {
        return nil, ErrSKUNotFound
    }
    if err != nil {
        return nil, fmt.Errorf("query current: %w", err)
    }

    // Check if enough stock
    available := current.Quantity - current.Reserved
    if available < req.Quantity {
        s.metrics.IncInsufficientStock(req.SKU)
        
        // Log for analytics
        s.logger.Warn("insufficient stock",
            "sku", req.SKU,
            "requested", req.Quantity,
            "available", available,
            "user_id", req.UserID,
        )
        
        return &DecrementResult{
            Success:     false,
            NewQuantity: current.Quantity,
            Error:       ErrInsufficientStock,
        }, nil
    }

    // Layer 4: Optimistic lock update
    result, err := tx.ExecContext(ctx, `
        UPDATE inventory 
        SET 
            quantity = quantity - $1,
            reserved = reserved,
            version = version + 1,
            updated_at = NOW()
        WHERE 
            sku = $2 
            AND version = $3  -- optimistic lock
            AND quantity >= $1  -- protection against corrupted data
    `, req.Quantity, req.SKU, current.Version)
    
    if err != nil {
        return nil, fmt.Errorf("update: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return nil, fmt.Errorf("rows affected: %w", err)
    }

    if rows == 0 {
        // Version conflict - someone changed data between SELECT and UPDATE
        s.metrics.IncOptimisticLockConflict(req.SKU)
        
        return &DecrementResult{
            Success:   false,
            Error:     ErrOptimisticLockConflict,
            Retryable: true, // can retry
        }, nil
    }

    // Layer 5: Audit log (for compliance)
    _, err = tx.ExecContext(ctx, `
        INSERT INTO inventory_audit (
            sku, user_id, request_id, quantity_delta, 
            old_quantity, new_quantity, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, NOW())
    `, req.SKU, req.UserID, req.RequestID, -req.Quantity,
        current.Quantity, current.Quantity-req.Quantity)
    
    if err != nil {
        s.logger.Error("failed to write audit log", "error", err)
        // Don't rollback transaction, just log
    }

    // Layer 6: Commit transaction
    if err = tx.Commit(); err != nil {
        return nil, fmt.Errorf("commit: %w", err)
    }

    // Layer 7: Cache invalidation
    s.invalidateCache(ctx, req.SKU)

    // Success!
    s.metrics.IncSuccessfulDecrement(req.SKU, req.Quantity)
    
    return &DecrementResult{
        Success:     true,
        NewQuantity: current.Quantity - req.Quantity,
        OldVersion:  current.Version,
    }, nil
}
```

---

## 4. Error Handling and Retry Logic

### 4.1. Error Classification

```go
type ErrorClass int

const (
    ErrorClassRetryable ErrorClass = iota // can retry
    ErrorClassPermanent                    // retrying useless
    ErrorClassThrottled                    // too many requests
)

func classifyError(err error) ErrorClass {
    switch {
    case errors.Is(err, ErrOptimisticLockConflict):
        return ErrorClassRetryable
    case errors.Is(err, ErrInsufficientStock):
        return ErrorClassPermanent
    case errors.Is(err, context.DeadlineExceeded):
        return ErrorClassRetryable
    case errors.Is(err, ErrDeadlock):
        return ErrorClassRetryable
    default:
        return ErrorClassPermanent
    }
}
```

### 4.2. Retry Mechanism with Exponential Backoff

```go
type RetryConfig struct {
    MaxAttempts     int
    InitialInterval time.Duration
    MaxInterval     time.Duration
    Multiplier      float64
    Jitter          float64 // randomness to avoid thundering herd
}

func (s *InventoryService) WithRetry(ctx context.Context, req *DecrementRequest) (*DecrementResult, error) {
    config := &RetryConfig{
        MaxAttempts:     3,
        InitialInterval: 100 * time.Millisecond,
        MaxInterval:     2 * time.Second,
        Multiplier:      2.0,
        Jitter:          0.1,
    }

    var result *DecrementResult
    var err error

    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        // Check context before each attempt
        if err := ctx.Err(); err != nil {
            return nil, fmt.Errorf("context cancelled: %w", err)
        }

        result, err = s.AtomicDecrement(ctx, req)
        
        if err == nil {
            if result.Success {
                return result, nil
            }
            // Successful request but failed operation (insufficient stock)
            if !result.Retryable {
                return result, nil
            }
        }

        // Classify error
        errorClass := classifyError(err)
        if errorClass == ErrorClassPermanent {
            return result, err
        }

        // Calculate delay with jitter
        backoff := s.calculateBackoff(attempt, config)
        
        s.logger.Debug("retrying operation",
            "attempt", attempt+1,
            "max_attempts", config.MaxAttempts,
            "backoff", backoff,
            "error", err,
        )

        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(backoff):
            continue
        }
    }

    return nil, fmt.Errorf("max retry attempts exceeded: %w", err)
}

func (s *InventoryService) calculateBackoff(attempt int, config *RetryConfig) time.Duration {
    if attempt == 0 {
        return config.InitialInterval
    }

    // Exponential growth
    interval := float64(config.InitialInterval) * math.Pow(config.Multiplier, float64(attempt))
    if interval > float64(config.MaxInterval) {
        interval = float64(config.MaxInterval)
    }

    // Add jitter to avoid thundering herd
    jitter := interval * config.Jitter * (rand.Float64()*2 - 1)
    
    return time.Duration(interval + jitter)
}
```

---

## 5. Monitoring and Metrics

### 5.1. Key Metrics

```go
type InventoryMetrics struct {
    // Operation counters
    decrementsTotal *prometheus.CounterVec
    successesTotal  *prometheus.CounterVec
    failuresTotal   *prometheus.CounterVec
    
    // Business metrics
    stockLevel      *prometheus.GaugeVec
    reservedLevel   *prometheus.GaugeVec
    
    // Technical metrics
    operationDuration *prometheus.HistogramVec
    lockContention    *prometheus.CounterVec
    retryAttempts     *prometheus.HistogramVec
    
    // Race condition metrics
    optimisticLockConflicts *prometheus.CounterVec
    deadlocks              *prometheus.CounterVec
}

func NewInventoryMetrics(reg *prometheus.Registry) *InventoryMetrics {
    m := &InventoryMetrics{
        decrementsTotal: promauto.With(reg).NewCounterVec(
            prometheus.CounterOpts{
                Name: "inventory_decrements_total",
                Help: "Total number of decrement attempts",
            },
            []string{"sku", "result"}, // result: success, insufficient, error
        ),
        
        optimisticLockConflicts: promauto.With(reg).NewCounterVec(
            prometheus.CounterOpts{
                Name: "inventory_optimistic_lock_conflicts_total",
                Help: "Number of optimistic lock conflicts",
            },
            []string{"sku"},
        ),
        
        operationDuration: promauto.With(reg).NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "inventory_operation_duration_seconds",
                Help:    "Duration of inventory operations",
                Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
            },
            []string{"operation", "result"},
        ),
        
        stockLevel: promauto.With(reg).NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "inventory_stock_level",
                Help: "Current stock level by SKU",
            },
            []string{"sku"},
        ),
    }
    
    // Start periodic stock level updates
    go m.periodicallyUpdateStockLevels(context.Background(), 30*time.Second)
    
    return m
}
```

### 5.2. Structured Logging

```go
type InventoryLog struct {
    Level       string    `json:"level"`
    Timestamp   time.Time `json:"@timestamp"`
    Operation   string    `json:"operation"` // "decrement", "reserve", "restore"
    
    // Context
    SKU         string    `json:"sku"`
    RequestID   string    `json:"request_id"`
    UserID      string    `json:"user_id,omitempty"`
    
    // Operation data
    Quantity    int       `json:"quantity"`
    BeforeQty   int       `json:"before_quantity"`
    AfterQty    int       `json:"after_quantity"`
    Version     int64     `json:"version"`
    
    // Result
    Success     bool      `json:"success"`
    Error       string    `json:"error,omitempty"`
    Retryable   bool      `json:"retryable,omitempty"`
    
    // Performance
    Duration    string    `json:"duration_ms"`
    RetryCount  int       `json:"retry_count,omitempty"`
}

func (s *InventoryService) logOperation(ctx context.Context, log *InventoryLog) {
    // Always add trace information
    log.Timestamp = time.Now()
    
    if span := trace.SpanFromContext(ctx); span != nil {
        log.RequestID = span.SpanContext().TraceID().String()
    }
    
    // Determine log level
    level := slog.LevelInfo
    if log.Error != "" {
        level = slog.LevelError
    }
    
    s.logger.Log(ctx, level, "inventory operation",
        "sku", log.SKU,
        "operation", log.Operation,
        "success", log.Success,
        "duration", log.Duration,
        "error", log.Error,
    )
}
```

### 5.3. Alerts

```yaml
alerts:
  - name: "High Optimistic Lock Conflict Rate"
    condition: "rate(inventory_optimistic_lock_conflicts_total[5m]) > 10"
    severity: "warning"
    description: "High contention on inventory updates"
    summary: "Too many retries due to concurrent updates"
    
  - name: "Stock Level Critical"
    condition: "inventory_stock_level{sku='POPULAR_ITEM'} < 10"
    severity: "critical"
    description: "Popular item running out of stock"
    
  - name: "High Operation Latency"
    condition: "histogram_quantile(0.95, rate(inventory_operation_duration_bucket[5m])) > 0.5"
    severity: "warning"
    description: "Inventory operations are slow"
    
  - name: "Multiple Deadlocks"
    condition: "increase(inventory_deadlocks_total[5m]) > 0"
    severity: "critical"
    description: "Database deadlocks detected in inventory operations"
```

---

## 6. Testing

### 6.1. Race Condition Test

```go
func TestInventory_ConcurrentDecrements(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping race test in short mode")
    }

    // Setup
    db := setupTestDB(t)
    service := NewInventoryService(db, redis.NewClient(&redis.Options{}))
    
    ctx := context.Background()
    sku := "TEST-SKU-001"
    initialQty := 100
    
    // Initialize inventory
    err := service.InitInventory(ctx, sku, initialQty)
    require.NoError(t, err)

    // Run concurrent requests
    const numRequests = 1000
    const decrementQty = 1
    
    var wg sync.WaitGroup
    results := make(chan *DecrementResult, numRequests)
    
    for i := 0; i < numRequests; i++ {
        wg.Add(1)
        go func(requestID int) {
            defer wg.Done()
            
            result, err := service.AtomicDecrement(ctx, &DecrementRequest{
                SKU:       sku,
                Quantity:  decrementQty,
                RequestID: fmt.Sprintf("req-%d", requestID),
                UserID:    "test-user",
            })
            
            if err != nil {
                t.Logf("Request %d failed: %v", requestID, err)
                results <- &DecrementResult{Success: false, Error: err}
            } else {
                results <- result
            }
        }(i)
    }
    
    // Wait for all goroutines
    wg.Wait()
    close(results)

    // Analyze results
    var successCount int
    var insufficientCount int
    var conflictCount int
    
    for res := range results {
        if res.Success {
            successCount++
        } else if errors.Is(res.Error, ErrInsufficientStock) {
            insufficientCount++
        } else if errors.Is(res.Error, ErrOptimisticLockConflict) {
            conflictCount++
        }
    }

    // Check invariants
    finalInventory, err := service.GetInventory(ctx, sku)
    require.NoError(t, err)
    
    // Total decremented items should equal initial - final
    assert.Equal(t, initialQty, finalInventory.Quantity+successCount)
    
    // Never go negative
    assert.GreaterOrEqual(t, finalInventory.Quantity, 0)
    
    // Success + insufficient should equal total requests
    // (conflicts are retries, they shouldn't decrement counter)
    assert.Equal(t, numRequests, successCount+insufficientCount)
    
    t.Logf("Results: %d success, %d insufficient, %d conflicts",
        successCount, insufficientCount, conflictCount)
}
```

### 6.2. Load Testing

```go
func BenchmarkInventory_Concurrent(b *testing.B) {
    db := setupBenchDB(b)
    service := NewInventoryService(db, redis.NewClient(&redis.Options{}))
    
    sku := "BENCH-SKU"
    service.InitInventory(context.Background(), sku, 1000000)
    
    b.RunParallel(func(pb *testing.PB) {
        requestID := 0
        for pb.Next() {
            requestID++
            _, err := service.AtomicDecrement(context.Background(), &DecrementRequest{
                SKU:       sku,
                Quantity:  1,
                RequestID: fmt.Sprintf("bench-%d", requestID),
                UserID:    "bench-user",
            })
            if err != nil && !errors.Is(err, ErrInsufficientStock) {
                b.Errorf("Unexpected error: %v", err)
            }
        }
    })
}
```

### 6.3. Isolation and Integrity Tests

```go
func TestInventory_TransactionIsolation(t *testing.T) {
    // Test for phantom reads
    t.Run("phantom read", func(t *testing.T) {
        // Verify REPEATABLE READ prevents phantoms
    })
    
    // Test for non-repeatable reads
    t.Run("non-repeatable read", func(t *testing.T) {
        // Verify SELECT FOR UPDATE locks the row
    })
    
    // Test for lost updates
    t.Run("lost update", func(t *testing.T) {
        // Verify parallel updates don't lose changes
    })
    
    // Test for deadlock
    t.Run("deadlock detection", func(t *testing.T) {
        // Verify deadlock leads to error, not hang
    })
}
```

---

## 7. Production Checklist

### ✅ Must have (critical for production)
- [ ] **Never go negative**: CHECK constraint at database level
- [ ] **Atomic operations**: Transactions with proper isolation level
- [ ] **Race condition protection**: `SELECT FOR UPDATE` or version column
- [ ] **Deadlock detection and retry**: Detect and retry deadlocks
- [ ] **Timeouts on all operations**: Don't wait forever for locks
- [ ] **Metrics**: latency, success rate, conflict rate
- [ ] **Logging**: Audit all changes
- [ ] **Graceful degradation**: Don't crash if Redis is down

### ⚠️ Should have (important for production readiness)
- [ ] **Idempotency**: Protection against duplicate requests
- [ ] **Circuit breaker**: Protection against cascading failures
- [ ] **Rate limiting**: DoS protection on inventory endpoints
- [ ] **Integration tests**: Tests with real DB under load
- [ ] **Documentation**: API contracts, usage examples
- [ ] **Reservation mechanism**: reservation + confirmation pattern

### 🚀 Nice to have (enterprise level)
- [ ] **Distributed transactions**: SAGA pattern for跨-service operations
- [ ] **Predictive analytics**: ML models for shortage prediction
- [ ] **Dynamic pricing**: Price changes during scarcity
- [ ] **A/B testing**: Different locking strategies
- [ ] **Auto-scaling**: During peak loads

---

## 8. Common Mistakes and Solutions

### 8.1. Mistake: Using `SELECT ... FOR UPDATE` Without Index
```sql
-- BAD: Locks entire table
SELECT * FROM inventory WHERE sku = 'test' FOR UPDATE;

-- GOOD: Uses index, locks only needed row
-- Need index: CREATE INDEX idx_inventory_sku ON inventory(sku);
SELECT * FROM inventory WHERE sku = 'test' FOR UPDATE;
```

### 8.2. Mistake: Long Transaction Holding Locks
```go
// BAD: Hold lock during external call
func (s *InventoryService) BadDecrement(ctx context.Context, sku string) error {
    tx, _ := s.db.Begin()
    defer tx.Rollback()
    
    // Lock the row
    var qty int
    tx.QueryRow("SELECT quantity FROM inventory WHERE sku = $1 FOR UPDATE", sku).Scan(&qty)
    
    // MAKING EXTERNAL CALL (HTTP, Kafka) - LOCK IS HELD!
    err := s.paymentService.Charge(ctx, ...)
    if err != nil {
        return err
    }
    
    tx.Exec("UPDATE inventory SET quantity = $1 WHERE sku = $2", qty-1, sku)
    return tx.Commit()
}

// GOOD: Lock only during update
func (s *InventoryService) GoodDecrement(ctx context.Context, sku string) error {
    // Make external calls first
    err := s.paymentService.Charge(ctx, ...)
    if err != nil {
        return err
    }
    
    // Then short transaction
    return s.AtomicDecrement(ctx, sku, 1)
}
```

### 8.3. Mistake: Wrong Isolation Level
```sql
-- BAD: READ COMMITTED can have phantoms
SET TRANSACTION ISOLATION LEVEL READ COMMITTED;

-- GOOD: REPEATABLE READ or SERIALIZABLE for inventory
SET TRANSACTION ISOLATION LEVEL REPEATABLE READ;
```

### 8.4. Mistake: Unhandled Deadlock
```go
// BAD: Deadlock leads to 500 error
func (s *InventoryService) BadDecrement() {
    _, err := db.Exec("UPDATE inventory SET quantity = quantity - 1 WHERE sku = $1", sku)
    if err != nil {
        // Just return 500
        return err
    }
}

// GOOD: Detect deadlock and retry
func (s *InventoryService) GoodDecrement() error {
    for attempts := 0; attempts < 3; attempts++ {
        _, err := db.Exec("UPDATE inventory SET quantity = quantity - 1 WHERE sku = $1", sku)
        if err == nil {
            return nil
        }
        
        // Check deadlock error code (PostgreSQL: 40P01)
        if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "40P01" {
            time.Sleep(time.Duration(attempts*100) * time.Millisecond)
            continue
        }
        return err
    }
    return fmt.Errorf("max deadlock retry attempts exceeded")
}
```

---

## Conclusion

Atomic inventory is a classic concurrency problem where you need to balance between:

1. **Data integrity** — never sell more than available
2. **Performance** — handle thousands of requests per second
3. **User experience** — give everyone a fair chance

Key takeaways:
- **Never trust a single protection layer** — use multi-layer approach
- **Measure everything** — latency, conflicts, deadlocks
- **Test under load** — race conditions only appear under contention
- **Have a plan B** — when in doubt, block rather than allow incorrect operations
- **Think about business** — technical solution must meet business requirements
