package consensus

import (
	"time"
)

type OutageScope string

const (
	ScopeGlobal    OutageScope = "GLOBAL_OUTAGE"
	ScopeLocalized OutageScope = "LOCALIZED_EDGE_OUTAGE"
	ScopeHealthy   OutageScope = "HEALTHY"
)

type RegionProbeStatus struct {
	Region     string        `json:"region"`
	IsUp       bool          `json:"is_up"`
	Latency    time.Duration `json:"latency"`
	HTTPStatus int           `json:"http_status"`
	Error      string        `json:"error,omitempty"`
}

type GeoConsensusEvaluation struct {
	TargetID         string              `json:"target_id"`
	Scope            OutageScope         `json:"scope"`
	FailingRegions   []string            `json:"failing_regions"`
	HealthyRegions   []string            `json:"healthy_regions"`
	TotalRegions     int                 `json:"total_regions"`
	FailureRatio     float64             `json:"failure_ratio"`
	AverageLatency   time.Duration       `json:"average_latency"`
	EvaluatedAt      time.Time           `json:"evaluated_at"`
}

type GeoMatrixEvaluator struct {
	GlobalThreshold float64 // e.g. 0.60 (60% or more failing -> Global)
}

func NewGeoMatrixEvaluator(threshold float64) *GeoMatrixEvaluator {
	if threshold <= 0 || threshold > 1.0 {
		threshold = 0.60
	}
	return &GeoMatrixEvaluator{GlobalThreshold: threshold}
}

func (e *GeoMatrixEvaluator) Evaluate(targetID string, probes []RegionProbeStatus) GeoConsensusEvaluation {
	if len(probes) == 0 {
		return GeoConsensusEvaluation{
			TargetID:    targetID,
			Scope:       ScopeHealthy,
			EvaluatedAt: time.Now().UTC(),
		}
	}

	var failing []string
	var healthy []string
	var totalLatency time.Duration

	for _, p := range probes {
		totalLatency += p.Latency
		if p.IsUp {
			healthy = append(healthy, p.Region)
		} else {
			failing = append(failing, p.Region)
		}
	}

	failureRatio := float64(len(failing)) / float64(len(probes))
	avgLatency := totalLatency / time.Duration(len(probes))

	var scope OutageScope
	switch {
	case failureRatio >= e.GlobalThreshold:
		scope = ScopeGlobal
	case len(failing) > 0:
		scope = ScopeLocalized
	default:
		scope = ScopeHealthy
	}

	return GeoConsensusEvaluation{
		TargetID:       targetID,
		Scope:          scope,
		FailingRegions: failing,
		HealthyRegions: healthy,
		TotalRegions:   len(probes),
		FailureRatio:   failureRatio,
		AverageLatency: avgLatency,
		EvaluatedAt:    time.Now().UTC(),
	}
}
