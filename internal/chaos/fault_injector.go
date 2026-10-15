package chaos

import (
	"context"
	"math/rand"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type FaultType string

const (
	FaultLatencySpike   FaultType = "LATENCY_SPIKE"
	FaultPacketLoss     FaultType = "PACKET_LOSS"
	FaultHTTPStatus500  FaultType = "HTTP_STATUS_500"
	FaultHTTPStatus503  FaultType = "HTTP_STATUS_503"
	FaultTimeout        FaultType = "TIMEOUT"
)

type FaultRule struct {
	ID          string        `json:"id"`
	TargetID    string        `json:"target_id"`
	Type        FaultType     `json:"type"`
	AddedDelay  time.Duration `json:"added_delay,omitempty"`
	Probability float64       `json:"probability"` // 0.0 to 1.0
	Active      bool          `json:"active"`
}

type FaultInjector struct {
	rules map[string]*FaultRule
}

func NewFaultInjector() *FaultInjector {
	return &FaultInjector{
		rules: make(map[string]*FaultRule),
	}
}

func (fi *FaultInjector) RegisterRule(rule *FaultRule) {
	fi.rules[rule.TargetID] = rule
}

func (fi *FaultInjector) RemoveRule(targetID string) {
	delete(fi.rules, targetID)
}

func (fi *FaultInjector) ApplyFault(ctx context.Context, target *models.Target, originalResult *models.CheckResult) (*models.CheckResult, bool) {
	rule, exists := fi.rules[target.ID]
	if !exists || !rule.Active {
		return originalResult, false
	}

	if rand.Float64() > rule.Probability {
		return originalResult, false
	}

	mutated := *originalResult

	switch rule.Type {
	case FaultLatencySpike:
		mutated.Latency += rule.AddedDelay
		select {
		case <-time.After(rule.AddedDelay):
		case <-ctx.Done():
		}
		return &mutated, true

	case FaultPacketLoss:
		mutated.Status = models.StatusDown
		mutated.ErrorMessage = "simulated chaos: packet loss 100% on interface"
		return &mutated, true

	case FaultHTTPStatus500:
		mutated.Status = models.StatusDown
		mutated.ErrorMessage = "simulated chaos: HTTP 500 Internal Server Error"
		return &mutated, true

	case FaultHTTPStatus503:
		mutated.Status = models.StatusDown
		mutated.ErrorMessage = "simulated chaos: HTTP 503 Service Unavailable"
		return &mutated, true

	case FaultTimeout:
		mutated.Status = models.StatusDown
		mutated.Latency = target.Timeout
		mutated.ErrorMessage = "simulated chaos: request context deadline exceeded"
		return &mutated, true
	}

	return originalResult, false
}
