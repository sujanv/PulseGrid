package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/sujanv/PulseGrid/internal/models"
)

var ErrPoolFull = errors.New("worker pool buffer is full (backpressure)")

type Job struct {
	Target   *models.Target
	Execute  func(ctx context.Context, target *models.Target) (*models.CheckResult, error)
	Callback func(result *models.CheckResult, err error)
}

type WorkerPool struct {
	maxWorkers    int32
	activeWorkers int32
	jobs          chan Job
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	stopped       int32
}

func NewWorkerPool(workerCount, queueCapacity int) *WorkerPool {
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
		return errors.New("worker pool is stopped")
	}

	select {
	case wp.jobs <- job:
		return nil
	default:
		return ErrPoolFull
	}
}
