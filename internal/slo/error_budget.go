package slo

import (
	"time"
)

type ErrorBudget struct {
	TargetID                 string        `json:"target_id"`
	SLOObjective             float64       `json:"slo_objective"`
	PeriodDays               int           `json:"period_days"`
	TotalMinutes             float64       `json:"total_minutes"`
	AllowedDowntimeMinutes   float64       `json:"allowed_downtime_minutes"`
	ConsumedDowntimeMinutes  float64       `json:"consumed_downtime_minutes"`
	RemainingDowntimeMinutes float64       `json:"remaining_downtime_minutes"`
	BudgetRemainingPercent   float64       `json:"budget_remaining_percent"`
	TimeToExhaustionHours    float64       `json:"time_to_exhaustion_hours"`
	IsDepleted               bool          `json:"is_depleted"`
}

type ErrorBudgetCalculator struct{}

func NewErrorBudgetCalculator() *ErrorBudgetCalculator {
	return &ErrorBudgetCalculator{}
}

func (c *ErrorBudgetCalculator) Calculate(targetID string, sloObjective float64, periodDays int, consumedDowntime time.Duration, currentBurnRate float64) ErrorBudget {
	if periodDays <= 0 {
		periodDays = 30
	}
	totalMinutes := float64(periodDays * 24 * 60)
	allowedDowntime := totalMinutes * (1.0 - (sloObjective / 100.0))
	consumedMinutes := consumedDowntime.Minutes()
	remainingMinutes := allowedDowntime - consumedMinutes
	if remainingMinutes < 0 {
		remainingMinutes = 0
	}

	budgetRemainingPercent := 0.0
	if allowedDowntime > 0 {
		budgetRemainingPercent = (remainingMinutes / allowedDowntime) * 100.0
	}

	timeToExhaustion := 9999.0 // infinity placeholder
	if currentBurnRate > 0 && remainingMinutes > 0 {
		burnRatePerMinute := (allowedDowntime / totalMinutes) * currentBurnRate
		if burnRatePerMinute > 0 {
			timeToExhaustion = (remainingMinutes / burnRatePerMinute) / 60.0
		}
	} else if remainingMinutes == 0 {
		timeToExhaustion = 0.0
	}

	return ErrorBudget{
		TargetID:                 targetID,
		SLOObjective:             sloObjective,
		PeriodDays:               periodDays,
		TotalMinutes:             totalMinutes,
		AllowedDowntimeMinutes:   allowedDowntime,
		ConsumedDowntimeMinutes:  consumedMinutes,
		RemainingDowntimeMinutes: remainingMinutes,
		BudgetRemainingPercent:   budgetRemainingPercent,
		TimeToExhaustionHours:    timeToExhaustion,
		IsDepleted:               remainingMinutes <= 0,
	}
}
