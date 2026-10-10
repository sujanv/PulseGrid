package test

import (
	"context"
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/pool"
)

func BenchmarkWorkerPoolSubmission(b *testing.B) {
	p := pool.NewWorkerPool(50, 10000)
	defer p.Stop(2 * time.Second)

	job := pool.Job{
		Target: &models.Target{Name: "bench"},
		Execute: func(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
			return nil, nil
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = p.Submit(job)
	}
}

func BenchmarkTokenBucketRateLimiter(b *testing.B) {
	tb := pool.NewTokenBucket(100000, 100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tb.Allow()
	}
}
