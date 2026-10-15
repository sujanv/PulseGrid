package chaos

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type ExperimentStatus string

const (
	StatusScheduled ExperimentStatus = "SCHEDULED"
	StatusRunning   ExperimentStatus = "RUNNING"
	StatusCompleted ExperimentStatus = "COMPLETED"
	StatusAborted   ExperimentStatus = "ABORTED"
)

type ChaosExperiment struct {
	ID                 string           `json:"id"`
	Name               string           `json:"name"`
	TargetID           string           `json:"target_id"`
	TargetEnv          string           `json:"target_env"` // must be "staging" or "test"
	FaultRule          FaultRule        `json:"fault_rule"`
	Duration           time.Duration    `json:"duration"`
	MaxBlastRadiusPct  float64          `json:"max_blast_radius_pct"`
	Status             ExperimentStatus `json:"status"`
	StartedAt          *time.Time       `json:"started_at,omitempty"`
	EndedAt            *time.Time       `json:"ended_at,omitempty"`
}

type ChaosScheduler struct {
	mu          sync.Mutex
	injector    *FaultInjector
	experiments map[string]*ChaosExperiment
}

func NewChaosScheduler(injector *FaultInjector) *ChaosScheduler {
	return &ChaosScheduler{
		injector:    injector,
		experiments: make(map[string]*ChaosExperiment),
	}
}

func (s *ChaosScheduler) LaunchExperiment(exp *ChaosExperiment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Blast radius safety check: Never run chaos in production targets
	if exp.TargetEnv == "production" || exp.TargetEnv == "prod" {
		return errors.New("safety violation: chaos experiments are strictly prohibited in production targets")
	}

	if exp.MaxBlastRadiusPct > 50.0 {
		return fmt.Errorf("safety violation: max blast radius %.1f%% exceeds 50%% ceiling", exp.MaxBlastRadiusPct)
	}

	now := time.Now().UTC()
	exp.Status = StatusRunning
	exp.StartedAt = &now
	s.experiments[exp.ID] = exp

	rule := exp.FaultRule
	rule.Active = true
	s.injector.RegisterRule(&rule)

	// Auto-abort / complete after Duration
	go func() {
		time.Sleep(exp.Duration)
		s.StopExperiment(exp.ID, StatusCompleted)
	}()

	return nil
}

func (s *ChaosScheduler) StopExperiment(id string, status ExperimentStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if exp, ok := s.experiments[id]; ok && exp.Status == StatusRunning {
		now := time.Now().UTC()
		exp.Status = status
		exp.EndedAt = &now
		s.injector.RemoveRule(exp.TargetID)
	}
}

func (s *ChaosScheduler) GetExperiment(id string) (*ChaosExperiment, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.experiments[id]
	return exp, ok
}
