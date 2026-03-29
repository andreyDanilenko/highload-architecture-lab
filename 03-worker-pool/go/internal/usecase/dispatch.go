package usecase

import (
	"context"

	"worker-pool/internal/domain"
)

// DispatchOutcome is the result of accepting or rejecting a task.
type DispatchOutcome struct {
	Accepted  bool
	QueueFull bool
	Err       error
}

// TaskDispatcher is the port for enqueueing work (naive goroutine, bounded pool, etc.).
type TaskDispatcher interface {
	Dispatch(ctx context.Context, taskID domain.TaskID) DispatchOutcome
}

// Stopper releases background resources (worker goroutines).
type Stopper interface {
	Stop()
}
