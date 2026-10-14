package slo

import (
	"time"
)

type AlertSeverity string

const (
	SeverityCritical AlertSeverity = "CRITICAL" // Page on-call immediately (1h / 6h window)
	SeverityHigh     AlertSeverity = "HIGH"     // Page on-call
	SeverityTicket   AlertSeverity = "TICKET"   // Create support/engineering ticket (24h / 3d window)
	SeverityNone     AlertSeverity = "NONE"
)

type BurnRateThreshold struct {
	WindowDuration time.Duration `json:"window_duration"`
	BurnRateFactor float64       `json:"burn_rate_factor"`
	BudgetPercent  float64       `json:"budget_percent"` // Error budget consumed in window
	Severity       AlertSeverity `json:"severity"`
}

// Google SRE Book Chapter 5 recommended multi-window burn rate thresholds
var StandardSREThresholds = []BurnRateThreshold{
	{WindowDuration: 1 * time.Hour, BurnRateFactor: 14.4, BudgetPercent: 2.0, Severity: SeverityCritical},
	{WindowDuration: 6 * time.Hour, BurnRateFactor: 6.0, BudgetPercent: 5.0, Severity: SeverityHigh},
	{WindowDuration: 24 * time.Hour, BurnRateFactor: 3.0, BudgetPercent: 10.0, Severity: SeverityTicket},
	{WindowDuration: 72 * time.Hour, BurnRateFactor: 1.0, BudgetPercent: 10.0, Severity: SeverityTicket},
}

type BurnRateResult struct {
	TargetID         string        `json:"target_id"`
	SLOObjective     float64       `json:"slo_objective"` // e.g. 99.9%
	WindowDuration   time.Duration `json:"window_duration"`
	TotalRequests    int64         `json:"total_requests"`
	FailedRequests   int64         `json:"failed_requests"`
	CurrentErrorRate float64       `json:"current_error_rate"`
	CurrentBurnRate  float64       `json:"current_burn_rate"`
	AlertSeverity    AlertSeverity `json:"alert_severity"`
	IsBreaching      bool          `json:"is_breaching"`
}

type BurnRateEvaluator struct {
	Thresholds []BurnRateThreshold
}

func NewBurnRateEvaluator(thresholds []BurnRateThreshold) *BurnRateEvaluator {
	if len(thresholds) == 0 {
		thresholds = StandardSREThresholds
	}
	return &BurnRateEvaluator{Thresholds: thresholds}
}

func (e *BurnRateEvaluator) Evaluate(targetID string, sloObjective float64, window time.Duration, totalReq, failedReq int64) BurnRateResult {
	if totalReq <= 0 {
		return BurnRateResult{
			TargetID:       targetID,
			SLOObjective:   sloObjective,
			WindowDuration: window,
			AlertSeverity:  SeverityNone,
		}
	}

	errorBudget := 1.0 - (sloObjective / 100.0)
	if errorBudget <= 0 {
		errorBudget = 0.001
	}

	actualErrorRate := float64(failedReq) / float64(totalReq)
	burnRate := actualErrorRate / errorBudget

	sev := SeverityNone
	breaching := false

	for _, t := range e.Thresholds {
		if window <= t.WindowDuration && burnRate >= t.BurnRateFactor {
			sev = t.Severity
			breaching = true
			break
		}
	}

	return BurnRateResult{
		TargetID:         targetID,
		SLOObjective:     sloObjective,
		WindowDuration:   window,
		TotalRequests:    totalReq,
		FailedRequests:   failedReq,
		CurrentErrorRate: actualErrorRate,
		CurrentBurnRate:  burnRate,
		AlertSeverity:    sev,
		IsBreaching:      breaching,
	}
}
