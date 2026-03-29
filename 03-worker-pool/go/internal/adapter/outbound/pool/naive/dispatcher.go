package naive

import (
	"context"
	"log"
	"time"

	"worker-pool/internal/domain"
	"worker-pool/internal/usecase"
	"worker-pool/internal/work"
)

// Dispatcher spawns one goroutine per task (demo: unbounded concurrency).
// With FFmpegSegmentConfig, each task runs a real ffmpeg child process (~1s synthetic or captured segment).
type Dispatcher struct {
	simulate time.Duration
	ffmpeg   *work.FFmpegSegmentConfig

	rootCtx    context.Context
	cancelRoot context.CancelFunc
}

// NewDispatcher builds a naive dispatcher. If ff is non-nil, work is ffmpeg instead of sleep.
func NewDispatcher(simulate time.Duration, ff *work.FFmpegSegmentConfig) *Dispatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &Dispatcher{
		simulate:   simulate,
		ffmpeg:     ff,
		rootCtx:    ctx,
		cancelRoot: cancel,
	}
}

// Dispatch always accepts and runs work in a new goroutine.
func (d *Dispatcher) Dispatch(_ context.Context, taskID domain.TaskID) usecase.DispatchOutcome {
	id := string(taskID)
	go func() {
		ctx := d.rootCtx
		start := time.Now()
		if d.ffmpeg != nil {
			if err := work.RunFFmpegOrUpload(ctx, *d.ffmpeg, id); err != nil {
				if ctx.Err() != nil {
					log.Printf("[naive ffmpeg] canceled id=%s after %s", id, time.Since(start))
					return
				}
				log.Printf("[naive ffmpeg] error id=%s: %v", id, err)
				return
			}
			log.Printf("[naive ffmpeg] done id=%s in %s", id, time.Since(start))
			return
		}
		select {
		case <-ctx.Done():
			log.Printf("[naive] canceled id=%s", id)
			return
		case <-time.After(d.simulate):
		}
		log.Printf("[naive] task done id=%s", id)
	}()
	return usecase.DispatchOutcome{Accepted: true}
}

// Stop cancels the root context (in-flight simulate waits and ffmpeg runs stop best-effort).
func (d *Dispatcher) Stop() {
	if d.cancelRoot != nil {
		d.cancelRoot()
	}
}
