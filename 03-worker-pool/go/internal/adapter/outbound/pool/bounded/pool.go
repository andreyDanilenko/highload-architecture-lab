package bounded

import (
	"context"
	"log"
	"sync"
	"time"

	"worker-pool/internal/domain"
	"worker-pool/internal/usecase"
	"worker-pool/internal/work"
)

// Pool is a fixed worker count + bounded task queue.
type Pool struct {
	tasks    chan string
	wg       sync.WaitGroup
	mu       sync.Mutex
	stopped  bool
	simulate time.Duration
	ff       *work.FFmpegSegmentConfig

	workerCtx context.Context
	cancel    context.CancelFunc
}

// New creates a started pool. If ff is non-nil, workers run a real ffmpeg job per task instead of sleep.
func New(workers, queueSize int, simulate time.Duration, ff *work.FFmpegSegmentConfig) *Pool {
	if workers < 1 {
		workers = 1
	}
	if queueSize < 1 {
		queueSize = 1
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &Pool{
		tasks:     make(chan string, queueSize),
		simulate:  simulate,
		ff:        ff,
		workerCtx: ctx,
		cancel:    cancel,
	}
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	return p
}

func (p *Pool) worker(id int) {
	defer p.wg.Done()
	for taskID := range p.tasks {
		if p.ff != nil {
			if err := work.RunFFmpegOrUpload(p.workerCtx, *p.ff, taskID); err != nil {
				if p.workerCtx.Err() != nil {
					log.Printf("[bounded worker=%d] ffmpeg canceled id=%s", id, taskID)
				} else {
					log.Printf("[bounded worker=%d] ffmpeg error id=%s: %v", id, taskID, err)
				}
				continue
			}
			log.Printf("[bounded worker=%d] ffmpeg done id=%s", id, taskID)
			continue
		}
		select {
		case <-p.workerCtx.Done():
			return
		case <-time.After(p.simulate):
		}
		log.Printf("[bounded worker=%d] task done id=%s", id, taskID)
	}
}

// Dispatch enqueues a task; returns QueueFull if the buffer is full.
func (p *Pool) Dispatch(_ context.Context, taskID domain.TaskID) usecase.DispatchOutcome {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stopped {
		return usecase.DispatchOutcome{Err: errStopped}
	}
	id := string(taskID)
	select {
	case p.tasks <- id:
		return usecase.DispatchOutcome{Accepted: true}
	default:
		return usecase.DispatchOutcome{QueueFull: true}
	}
}

// Stop closes the task channel and waits for workers (call after HTTP server shutdown).
func (p *Pool) Stop() {
	p.mu.Lock()
	if !p.stopped {
		p.stopped = true
		close(p.tasks)
	}
	p.mu.Unlock()
	if p.ff != nil {
		p.cancel()
	}
	p.wg.Wait()
}
