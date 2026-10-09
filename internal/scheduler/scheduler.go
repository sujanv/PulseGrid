package scheduler

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/sujanv/PulseGrid/internal/checker"
	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/pool"
)

type TargetCallback func(target *models.Target, result *models.CheckResult, err error)

type Scheduler struct {
	pool        *pool.WorkerPool
	registry    *checker.Registry
	rateLimiter *pool.PerHostRateLimiter
	callback    TargetCallback
	jitterPct   float64

	mu      sync.RWMutex
	targets map[string]*models.Target
	timers  map[string]*time.Timer

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewScheduler(
	wp *pool.WorkerPool,
	reg *checker.Registry,
	limiter *pool.PerHostRateLimiter,
	jitterPct float64,
	callback TargetCallback,
) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		pool:        wp,
		registry:    reg,
		rateLimiter: limiter,
		callback:    callback,
		jitterPct:   jitterPct,
		targets:     make(map[string]*models.Target),
		timers:      make(map[string]*time.Timer),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (s *Scheduler) RegisterTarget(target *models.Target) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.targets[target.ID] = target
	s.scheduleNext(target)
}

func (s *Scheduler) UnregisterTarget(targetID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if timer, ok := s.timers[targetID]; ok {
		timer.Stop()
		delete(s.timers, targetID)
	}
	delete(s.targets, targetID)
}

func (s *Scheduler) calculateNextInterval(interval time.Duration) time.Duration {
	if s.jitterPct <= 0 {
		return interval
	}
	jitterDelta := float64(interval) * s.jitterPct
	offset := (rand.Float64()*2 - 1) * jitterDelta
	next := time.Duration(float64(interval) + offset)
	if next < 500*time.Millisecond {
		next = 500 * time.Millisecond
	}
	return next
}

func (s *Scheduler) scheduleNext(target *models.Target) {
	if !target.Active {
		return
	}

	delay := s.calculateNextInterval(target.Interval)
	timer := time.AfterFunc(delay, func() {
		s.dispatch(target)
	})
	s.timers[target.ID] = timer
}

func (s *Scheduler) dispatch(target *models.Target) {
	s.mu.RLock()
	current, exists := s.targets[target.ID]
	s.mu.RUnlock()

	if !exists || !current.Active {
		return
	}

	job := pool.Job{
		Target: current,
		Execute: func(ctx context.Context, t *models.Target) (*models.CheckResult, error) {
			if s.rateLimiter != nil {
				if err := s.rateLimiter.Wait(ctx, t.URL); err != nil {
					return nil, err
				}
			}
			return s.registry.Execute(ctx, t)
		},
		Callback: func(result *models.CheckResult, err error) {
			if s.callback != nil {
				s.callback(target, result, err)
			}
		},
	}

	if err := s.pool.Submit(job); err != nil {
		log.Printf("[Scheduler] Dropping check for %s: %v", target.Name, err)
	}

	s.mu.Lock()
	if _, ok := s.targets[target.ID]; ok {
		s.scheduleNext(current)
	}
	s.mu.Unlock()
}

func (s *Scheduler) TargetCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.targets)
}

func (s *Scheduler) Stop() {
	s.cancel()

	s.mu.Lock()
	for _, timer := range s.timers {
		timer.Stop()
	}
	s.timers = make(map[string]*time.Timer)
	s.mu.Unlock()
}
