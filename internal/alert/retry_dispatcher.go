package alert

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type RetryDispatcher struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

func NewRetryDispatcher(maxRetries int, baseDelay, maxDelay time.Duration) *RetryDispatcher {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if baseDelay <= 0 {
		baseDelay = 100 * time.Millisecond
	}
	if maxDelay <= 0 {
		maxDelay = 2 * time.Second
	}
	return &RetryDispatcher{
		MaxRetries: maxRetries,
		BaseDelay:  baseDelay,
		MaxDelay:   maxDelay,
	}
}

func (d *RetryDispatcher) ExecuteWithRetry(ctx context.Context, action func() error) error {
	var lastErr error

	for attempt := 0; attempt <= d.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		lastErr = action()
		if lastErr == nil {
			return nil
		}

		if attempt == d.MaxRetries {
			break
		}

		// Exponential backoff: base * 2^attempt + jitter
		backoff := d.BaseDelay * (1 << attempt)
		if backoff > d.MaxDelay {
			backoff = d.MaxDelay
		}
		jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
		sleepDur := backoff + jitter

		select {
		case <-time.After(sleepDur):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("dispatcher failed after %d retries: %w", d.MaxRetries, lastErr)
}
