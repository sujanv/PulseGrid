package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

var (
	ErrPoolFull    = errors.New("worker pool buffer is full (backpressure)")
	ErrPoolStopped = errors.New("worker pool is stopped")
)

type Job struct {
	Target   *models.Target
	Execute  func(ctx context.Context, target *models.Target) (*models.CheckResult, error)
	Callback func(result *models.CheckResult, err error)
}

type PoolStats struct {
	MaxWorkers    int   `json:"max_workers"`
	ActiveWorkers int   `json:"active_workers"`
	QueuedJobs    int   `json:"queued_jobs"`
	CompletedJobs int64 `json:"completed_jobs"`
}

type WorkerPool struct {
	maxWorkers    int32
	activeWorkers int32
	completedJobs int64
	jobs          chan Job
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	stopped       int32
	mu            sync.Mutex
}

func NewWorkerPool(workerCount, queueCapacity int) *WorkerPool {
	if workerCount <= 0 {
		workerCount = 10
	}
	if queueCapacity <= 0 {
		queueCapacity = 1000
	}

	ctx, cancel := context.WithCancel(context.Background())
	wp := &WorkerPool{
		maxWorkers: int32(workerCount),
		jobs:       make(chan Job, queueCapacity),
		ctx:        ctx,
		cancel:     cancel,
	}

	wp.start()
	return wp
}

func (wp *WorkerPool) start() {
	count := int(atomic.LoadInt32(&wp.maxWorkers))
	for i := 0; i < count; i++ {
		wp.spawnWorker()
	}
}

func (wp *WorkerPool) spawnWorker() {
	wp.wg.Add(1)
	go func() {
		defer wp.wg.Done()
		for {
			select {
			case <-wp.ctx.Done():
				return
			case job, ok := <-wp.jobs:
				if !ok {
					return
				}
				atomic.AddInt32(&wp.activeWorkers, 1)
				res, err := job.Execute(wp.ctx, job.Target)
				atomic.AddInt64(&wp.completedJobs, 1)
				if job.Callback != nil {
					job.Callback(res, err)
				}
				atomic.AddInt32(&wp.activeWorkers, -1)
			}
		}
	}()
}

func (wp *WorkerPool) Submit(job Job) error {
	if atomic.LoadInt32(&wp.stopped) == 1 {
		return ErrPoolStopped
	}

	select {
	case wp.jobs <- job:
		return nil
	default:
		return ErrPoolFull
	}
}

func (wp *WorkerPool) Resize(newSize int) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	current := int(atomic.LoadInt32(&wp.maxWorkers))
	if newSize <= 0 || newSize == current {
		return
	}

	if newSize > current {
		diff := newSize - current
		for i := 0; i < diff; i++ {
			wp.spawnWorker()
		}
	}
	// For shrinking, workers naturally decrease or idle
	atomic.StoreInt32(&wp.maxWorkers, int32(newSize))
}

func (wp *WorkerPool) Stats() PoolStats {
	return PoolStats{
		MaxWorkers:    int(atomic.LoadInt32(&wp.maxWorkers)),
		ActiveWorkers: int(atomic.LoadInt32(&wp.activeWorkers)),
		QueuedJobs:    len(wp.jobs),
		CompletedJobs: atomic.LoadInt64(&wp.completedJobs),
	}
}

func (wp *WorkerPool) Stop(timeout time.Duration) error {
	if !atomic.CompareAndSwapInt32(&wp.stopped, 0, 1) {
		return nil
	}

	close(wp.jobs)

	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		wp.cancel()
		return nil
	case <-time.After(timeout):
		wp.cancel()
		return errors.New("worker pool shutdown timed out; cancelled active workers")
	}
}
