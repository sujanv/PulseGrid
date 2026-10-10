package test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/pool"
)

func TestWorkerPoolBoundedConcurrency(t *testing.T) {
	maxWorkers := 5
	p := pool.NewWorkerPool(maxWorkers, 100)
	defer p.Stop(1 * time.Second)

	var runningWorkers int32
	var maxObserved int32

	jobCount := 30
	done := make(chan struct{}, jobCount)

	for i := 0; i < jobCount; i++ {
		err := p.Submit(pool.Job{
			Target: &models.Target{Name: "test"},
			Execute: func(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
				curr := atomic.AddInt32(&runningWorkers, 1)
				for {
					max := atomic.LoadInt32(&maxObserved)
					if curr <= max || atomic.CompareAndSwapInt32(&maxObserved, max, curr) {
						break
					}
				}
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt32(&runningWorkers, -1)
				done <- struct{}{}
				return nil, nil
			},
		})
		if err != nil {
			t.Fatalf("unexpected submit error: %v", err)
		}
	}

	for i := 0; i < jobCount; i++ {
		<-done
	}

	if maxObserved > int32(maxWorkers) {
		t.Errorf("expected max %d workers, but observed %d", maxWorkers, maxObserved)
	}
}

func TestWorkerPoolBackpressure(t *testing.T) {
	p := pool.NewWorkerPool(1, 2)
	defer p.Stop(500 * time.Millisecond)

	// Fill worker and queue
	blocker := make(chan struct{})
	for i := 0; i < 3; i++ {
		_ = p.Submit(pool.Job{
			Target: &models.Target{Name: "test"},
			Execute: func(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
				<-blocker
				return nil, nil
			},
		})
	}

	// Next submit should trigger backpressure
	err := p.Submit(pool.Job{
		Target: &models.Target{Name: "overflow"},
		Execute: func(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
			return nil, nil
		},
	})

	if err != pool.ErrPoolFull {
		t.Errorf("expected ErrPoolFull, got: %v", err)
	}
	close(blocker)
}
