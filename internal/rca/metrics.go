package rca

import (
	"time"
)

type ReliabilityImpactMetrics struct {
	IncidentID            string        `json:"incident_id"`
	MTTD                  time.Duration `json:"mttd"` // Mean Time to Detect
	MTTR                  time.Duration `json:"mttr"` // Mean Time to Resolve
	TotalDowntimeMinutes  float64       `json:"total_downtime_minutes"`
	ErrorBudgetBurnedPct  float64       `json:"error_budget_burned_pct"`
	SLABreached           bool          `json:"sla_breached"`
}

type ReliabilityMetricsCalculator struct{}

func NewReliabilityMetricsCalculator() *ReliabilityMetricsCalculator {
	return &ReliabilityMetricsCalculator{}
}

func (c *ReliabilityMetricsCalculator) Calculate(incidentID string, failureStarted, firstDetected, resolved time.Time, sloObjective float64) ReliabilityImpactMetrics {
	mttd := firstDetected.Sub(failureStarted)
	if mttd < 0 {
		mttd = 0
	}

	mttr := resolved.Sub(firstDetected)
	if mttr < 0 {
		mttr = 0
	}

	totalDowntime := resolved.Sub(failureStarted).Minutes()
	if totalDowntime < 0 {
		totalDowntime = 0
	}

	// 30-day window allowed downtime in minutes
	totalMonthMinutes := 30.0 * 24.0 * 60.0
	allowedDowntimeMinutes := totalMonthMinutes * (1.0 - (sloObjective / 100.0))
	burnedPct := 0.0
	if allowedDowntimeMinutes > 0 {
		burnedPct = (totalDowntime / allowedDowntimeMinutes) * 100.0
	}

	return ReliabilityImpactMetrics{
		IncidentID:           incidentID,
		MTTD:                 mttd,
		MTTR:                 mttr,
		TotalDowntimeMinutes: totalDowntime,
		ErrorBudgetBurnedPct: burnedPct,
		SLABreached:          burnedPct > 100.0,
	}
}
