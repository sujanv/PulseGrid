package remediation

import (
	"sync"
	"time"
)

type CircuitBreaker struct {
	mu           sync.Mutex
	lastExecuted map[string]time.Time
}

func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		lastExecuted: make(map[string]time.Time),
	}
}

func (cb *CircuitBreaker) Allow(targetID string, cooldown time.Duration) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	last, exists := cb.lastExecuted[targetID]
	now := time.Now()
	if !exists || now.Sub(last) >= cooldown {
		cb.lastExecuted[targetID] = now
		return true
	}
	return false
}
