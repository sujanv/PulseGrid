package sampling

import (
	"math/rand"
	"sync"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type SamplingDecision struct {
	ShouldSample bool    `json:"should_sample"`
	SampleRate   float64 `json:"sample_rate"` // 0.0 to 1.0
	Reason       string  `json:"reason"`
}

type AdaptiveSampler struct {
	mu                  sync.RWMutex
	baseRate            float64       // Baseline sampling rate for healthy targets (e.g. 0.20 = 20%)
	degradedThresholdMs time.Duration // Latency threshold to ramp up sampling
	targetFailureState  map[string]bool
}

func NewAdaptiveSampler(baseRate float64, degradedThreshold time.Duration) *AdaptiveSampler {
	if baseRate <= 0.0 || baseRate > 1.0 {
		baseRate = 0.20
	}
	if degradedThreshold <= 0 {
		degradedThreshold = 500 * time.Millisecond
	}
	return &AdaptiveSampler{
		baseRate:            baseRate,
		degradedThresholdMs: degradedThreshold,
		targetFailureState:  make(map[string]bool),
	}
}

func (s *AdaptiveSampler) Evaluate(targetID string, lastStatus models.CheckStatus, lastLatency time.Duration) SamplingDecision {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. If target is DOWN: always 100% sample rate for maximum incident visibility
	if lastStatus == models.StatusDown {
		s.targetFailureState[targetID] = true
		return SamplingDecision{
			ShouldSample: true,
			SampleRate:   1.0,
			Reason:       "target is failing; 100% full sampling enforced",
		}
	}

	// 2. If target is DEGRADED (high latency): full 100% sampling to capture anomaly
	if lastLatency > s.degradedThresholdMs {
		return SamplingDecision{
			ShouldSample: true,
			SampleRate:   1.0,
			Reason:       "target latency degraded; full diagnostic sampling activated",
		}
	}

	// 3. Target is healthy: use baseline rate to conserve bandwidth and storage
	s.targetFailureState[targetID] = false
	sampled := rand.Float64() <= s.baseRate

	return SamplingDecision{
		ShouldSample: sampled,
		SampleRate:   s.baseRate,
		Reason:       "steady-state green; adaptive downsampled",
	}
}
