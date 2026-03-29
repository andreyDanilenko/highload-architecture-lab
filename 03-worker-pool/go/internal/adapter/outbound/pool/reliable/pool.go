package reliable

import (
	"context"
	"log"
	"sync"
	"time"

	"worker-pool/internal/domain"
	"worker-pool/internal/usecase"
	"worker-pool/internal/work"
)

// Pool is a bounded worker pool with panic recovery in workers.
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

// New creates a started pool. If ff is non-nil, workers run ffmpeg per task instead of sleep.
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
		p.runTask(id, taskID)
	}
}

func (p *Pool) runTask(workerID int, taskID string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[reliable worker=%d] panic recovered id=%s: %v", workerID, taskID, r)
		}
	}()
	if p.ff != nil {
		if err := work.RunFFmpegOrUpload(p.workerCtx, *p.ff, taskID); err != nil {
			if p.workerCtx.Err() != nil {
				log.Printf("[reliable worker=%d] ffmpeg canceled id=%s", workerID, taskID)
				return
			}
			log.Printf("[reliable worker=%d] ffmpeg error id=%s: %v", workerID, taskID, err)
			return
		}
		log.Printf("[reliable worker=%d] ffmpeg done id=%s", workerID, taskID)
		return
	}
	select {
	case <-p.workerCtx.Done():
		return
	case <-time.After(p.simulate):
	}
	log.Printf("[reliable worker=%d] task done id=%s", workerID, taskID)
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

// Stop closes the task channel and waits for workers.
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
