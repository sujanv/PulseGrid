package test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/pool"
)

func BenchmarkWorkerPoolHighThroughput(b *testing.B) {
	wp := pool.NewWorkerPool(100, 50000)
	defer wp.Stop(2 * time.Second)

	var completed int64
	job := pool.Job{
		Target: &models.Target{Name: "stress-target"},
		Execute: func(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
			atomic.AddInt64(&completed, 1)
			return nil, nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wp.Submit(job)
	}
}
