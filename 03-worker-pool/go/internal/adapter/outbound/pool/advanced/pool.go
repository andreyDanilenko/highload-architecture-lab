package advanced

import (
	"context"
	"time"

	"worker-pool/internal/adapter/outbound/pool/bounded"
	"worker-pool/internal/domain"
	"worker-pool/internal/usecase"
	"worker-pool/internal/work"
)

// Pool wraps a bounded pool with separate sizing (placeholder for priorities / dynamic scaling).
type Pool struct {
	inner *bounded.Pool
}

// New creates the advanced pool backed by a bounded implementation.
func New(workers, queueSize int, simulate time.Duration, ff *work.FFmpegSegmentConfig) *Pool {
	return &Pool{inner: bounded.New(workers, queueSize, simulate, ff)}
}

// Dispatch forwards to the inner pool.
func (p *Pool) Dispatch(ctx context.Context, taskID domain.TaskID) usecase.DispatchOutcome {
	return p.inner.Dispatch(ctx, taskID)
}

// Stop stops the inner pool.
func (p *Pool) Stop() {
	p.inner.Stop()
}
