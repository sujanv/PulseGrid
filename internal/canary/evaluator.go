package canary

import (
	"fmt"
	"time"
)

type RolloutDecision string

const (
	DecisionPromote  RolloutDecision = "PROMOTE"
	DecisionHalt     RolloutDecision = "HALT"
	DecisionRollback RolloutDecision = "ROLLBACK"
)

type TargetMetricsSummary struct {
	TotalChecks int64         `json:"total_checks"`
	FailedChecks int64        `json:"failed_checks"`
	ErrorRate   float64       `json:"error_rate"` // 0.0 to 1.0
	P95Latency  time.Duration `json:"p95_latency"`
	P99Latency  time.Duration `json:"p99_latency"`
}

type CanaryEvaluationResult struct {
	BaselineID     string               `json:"baseline_id"`
	CanaryID       string               `json:"canary_id"`
	Decision       RolloutDecision      `json:"decision"`
	ConfidencePct  float64              `json:"confidence_pct"`
	Reason         string               `json:"reason"`
	BaselineStats  TargetMetricsSummary `json:"baseline_stats"`
	CanaryStats    TargetMetricsSummary `json:"canary_stats"`
	EvaluatedAt    time.Time            `json:"evaluated_at"`
}

type CanaryEvaluator struct {
	maxErrorDelta float64       // e.g. 0.02 (2% error rate diff triggers halt)
	maxLatencyDiff time.Duration // e.g. 100ms P95 latency regression
}

func NewCanaryEvaluator(maxErrorDelta float64, maxLatencyDiff time.Duration) *CanaryEvaluator {
	if maxErrorDelta <= 0 {
		maxErrorDelta = 0.01 // 1% threshold
	}
	if maxLatencyDiff <= 0 {
		maxLatencyDiff = 50 * time.Millisecond
	}
	return &CanaryEvaluator{
		maxErrorDelta:  maxErrorDelta,
		maxLatencyDiff: maxLatencyDiff,
	}
}

func (e *CanaryEvaluator) Evaluate(baselineID, canaryID string, baseStats, canStats TargetMetricsSummary) CanaryEvaluationResult {
	now := time.Now().UTC()

	// 1. Calculate error rate delta
	errDelta := canStats.ErrorRate - baseStats.ErrorRate
	if errDelta > e.maxErrorDelta {
		return CanaryEvaluationResult{
			BaselineID:    baselineID,
			CanaryID:      canaryID,
			Decision:      DecisionRollback,
			ConfidencePct: 98.5,
			Reason:        fmt.Sprintf("canary error rate (%.2f%%) exceeds baseline (%.2f%%) by %.2f%%", canStats.ErrorRate*100, baseStats.ErrorRate*100, errDelta*100),
			BaselineStats: baseStats,
			CanaryStats:   canStats,
			EvaluatedAt:   now,
		}
	}

	// 2. Latency regression check
	latencyDiff := canStats.P95Latency - baseStats.P95Latency
	if latencyDiff > e.maxLatencyDiff {
		return CanaryEvaluationResult{
			BaselineID:    baselineID,
			CanaryID:      canaryID,
			Decision:      DecisionHalt,
			ConfidencePct: 92.0,
			Reason:        fmt.Sprintf("canary P95 latency (%v) regressed by %v over baseline (%v)", canStats.P95Latency, latencyDiff, baseStats.P95Latency),
			BaselineStats: baseStats,
			CanaryStats:   canStats,
			EvaluatedAt:   now,
		}
	}

	return CanaryEvaluationResult{
		BaselineID:    baselineID,
		CanaryID:      canaryID,
		Decision:      DecisionPromote,
		ConfidencePct: 99.0,
		Reason:        "canary matches or exceeds baseline performance criteria",
		BaselineStats: baseStats,
		CanaryStats:   canStats,
		EvaluatedAt:   now,
	}
}
