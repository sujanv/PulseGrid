package pool

import (
	"context"
	"net/url"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   float64
	tokens     float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		if tb.Allow() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}

type PerHostRateLimiter struct {
	mu         sync.RWMutex
	buckets    map[string]*TokenBucket
	rate       float64
	capacity   float64
	lastAccess map[string]time.Time
}

func NewPerHostRateLimiter(ratePerSecond float64) *PerHostRateLimiter {
	limiter := &PerHostRateLimiter{
		buckets:    make(map[string]*TokenBucket),
		rate:       ratePerSecond,
		capacity:   ratePerSecond * 2, // 2-second burst capacity
		lastAccess: make(map[string]time.Time),
	}

	go limiter.cleanupLoop()
	return limiter
}

func (l *PerHostRateLimiter) extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	return u.Hostname()
}

func (l *PerHostRateLimiter) getBucket(host string) *TokenBucket {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.lastAccess[host] = time.Now()
	bucket, ok := l.buckets[host]
	if !ok {
		bucket = NewTokenBucket(l.capacity, l.rate)
		l.buckets[host] = bucket
	}
	return bucket
}

func (l *PerHostRateLimiter) Allow(targetURL string) bool {
	host := l.extractHost(targetURL)
	bucket := l.getBucket(host)
	return bucket.Allow()
}

func (l *PerHostRateLimiter) Wait(ctx context.Context, targetURL string) error {
	host := l.extractHost(targetURL)
	bucket := l.getBucket(host)
	return bucket.Wait(ctx)
}

func (l *PerHostRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for host, lastTime := range l.lastAccess {
			if now.Sub(lastTime) > 30*time.Minute {
				delete(l.buckets, host)
				delete(l.lastAccess, host)
			}
		}
		l.mu.Unlock()
	}
}
