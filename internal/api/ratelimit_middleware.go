package api

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   float64
	tokens     float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

type TokenBucketLimiter struct {
	mu         sync.Mutex
	capacity   float64
	refillRate float64
	buckets    map[string]*TokenBucket
}

func NewTokenBucketLimiter(capacity, refillRate float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		capacity:   capacity,
		refillRate: refillRate,
		buckets:    make(map[string]*TokenBucket),
	}
}

func (l *TokenBucketLimiter) Allow(key string) (bool, int, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, exists := l.buckets[key]
	if !exists {
		b = &TokenBucket{
			capacity:   l.capacity,
			tokens:     l.capacity,
			refillRate: l.refillRate,
			lastRefill: now,
		}
		l.buckets[key] = b
	}

	// Refill tokens based on elapsed duration
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefill = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		rem := int(b.tokens)
		return true, rem, 0
	}

	missing := 1.0 - b.tokens
	retryAfter := time.Duration(missing/b.refillRate*float64(time.Second)) + 100*time.Millisecond
	return false, 0, retryAfter
}

func (l *TokenBucketLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		allowed, remaining, retryAfter := l.Allow(clientIP)

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(int(l.capacity)))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
			http.Error(w, "Too Many Requests: rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
