package alert

import (
	"context"
	"sync"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type EscalationTier struct {
	Level    int           `json:"level"`
	Delay    time.Duration `json:"delay"` // e.g. 0m (immediate), 10m, 30m
	Channels []string      `json:"channels"` // Channel IDs or Types
}

type EscalationPolicy struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	TargetIDs []string         `json:"target_ids"`
	Tiers     []EscalationTier `json:"tiers"`
}

type EscalationManager struct {
	mu         sync.RWMutex
	policies   map[string]*EscalationPolicy
	dispatcher *Dispatcher
	activeRuns map[string]*escalationRun
}

type escalationRun struct {
	incidentID string
	event      *models.AlertEvent
	cancel     context.CancelFunc
}

func NewEscalationManager(d *Dispatcher) *EscalationManager {
	return &EscalationManager{
		policies:   make(map[string]*EscalationPolicy),
		dispatcher: d,
		activeRuns: make(map[string]*escalationRun),
	}
}

func (m *EscalationManager) RegisterPolicy(p *EscalationPolicy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.policies[p.ID] = p
}

func (m *EscalationManager) HandleOutage(event *models.AlertEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Cancel existing run if any
	if run, exists := m.activeRuns[event.TargetID]; exists {
		run.cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.activeRuns[event.TargetID] = &escalationRun{
		incidentID: event.IncidentID,
		event:      event,
		cancel:     cancel,
	}

	go m.executeTiers(ctx, event)
}

func (m *EscalationManager) HandleRecovery(event *models.AlertEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if run, exists := m.activeRuns[event.TargetID]; exists {
		run.cancel()
		delete(m.activeRuns, event.TargetID)
	}

	// Always dispatch recovery immediately
	m.dispatcher.Dispatch(event)
}

func (m *EscalationManager) executeTiers(ctx context.Context, event *models.AlertEvent) {
	// Find matching policy or use default 2-tier escalation
	tiers := []EscalationTier{
		{Level: 1, Delay: 0},
		{Level: 2, Delay: 10 * time.Minute},
	}

	for _, tier := range tiers {
		if tier.Delay > 0 {
			select {
			case <-ctx.Done():
				return // outage was resolved or cancelled
			case <-time.After(tier.Delay):
			}
		}

		// Dispatch alert for this tier
		tierEvent := *event
		tierEvent.Message = fmt.Sprintf("[Escalation Level %d] %s", tier.Level, event.Message)
		m.dispatcher.Dispatch(&tierEvent)
	}
}
