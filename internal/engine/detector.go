package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/maintenance"
	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/storage/postgres"
)

type IncidentListener func(event *models.AlertEvent)

type TargetState struct {
	Status              models.CheckStatus
	ConsecutiveFailures int
	ConsecutiveSuccess  int
	CurrentIncident     *models.Incident
	LastCheckTime       time.Time
}

type IncidentDetector struct {
	incidentRepo       postgres.IncidentRepository
	listener           IncidentListener
	maintenanceManager *maintenance.Manager
	mu                 sync.RWMutex
	states             map[string]*TargetState
}

func NewIncidentDetector(repo postgres.IncidentRepository, listener IncidentListener) *IncidentDetector {
	return &IncidentDetector{
		incidentRepo: repo,
		listener:     listener,
		states:       make(map[string]*TargetState),
	}
}

func (d *IncidentDetector) SetMaintenanceManager(m *maintenance.Manager) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.maintenanceManager = m
}

func (d *IncidentDetector) ProcessResult(ctx context.Context, target *models.Target, res *models.CheckResult) (*models.Incident, *models.AlertEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if target is currently under scheduled maintenance
	if d.maintenanceManager != nil && d.maintenanceManager.IsUnderMaintenance(target.ID, res.Timestamp) {
		res.Status = models.StatusDegraded
		res.ErrorMessage = "[Maintenance Window Active] Alerts and downtime suppressed"
		return nil, nil
	}

	state, ok := d.states[target.ID]
	if !ok {
		state = &TargetState{
			Status: models.StatusUp,
		}
		d.states[target.ID] = state
	}
	state.LastCheckTime = res.Timestamp

	var generatedIncident *models.Incident
	var alert *models.AlertEvent

	if res.Status == models.StatusDown {
		state.ConsecutiveFailures++
		state.ConsecutiveSuccess = 0

		if state.ConsecutiveFailures >= target.FailureThreshold && state.CurrentIncident == nil {
			state.Status = models.StatusDown
			rootCause := res.ErrorMessage
			if rootCause == "" && res.StatusCode != 0 {
				rootCause = fmt.Sprintf("HTTP %d response", res.StatusCode)
			}

			inc := &models.Incident{
				ID:                  uuid.New().String(),
				TargetID:            target.ID,
				TargetName:          target.Name,
				StartedAt:           res.Timestamp,
				Status:              models.IncidentOpen,
				RootCause:           rootCause,
				ConsecutiveFailures: state.ConsecutiveFailures,
				Severity:            models.SeverityCritical,
			}
			state.CurrentIncident = inc
			generatedIncident = inc

			if d.incidentRepo != nil {
				_ = d.incidentRepo.CreateIncident(ctx, inc)
			}

			alert = &models.AlertEvent{
				ID:         uuid.New().String(),
				IncidentID: inc.ID,
				TargetID:   target.ID,
				TargetName: target.Name,
				TargetURL:  target.URL,
				EventType:  models.AlertOutage,
				Message:    fmt.Sprintf("Target %s is DOWN: %s", target.Name, rootCause),
				LatencyMs:  float64(res.Latency.Milliseconds()),
				StatusCode: res.StatusCode,
				Timestamp:  res.Timestamp,
			}
		}
	} else {
		state.ConsecutiveSuccess++
		state.ConsecutiveFailures = 0

		if state.CurrentIncident != nil && state.ConsecutiveSuccess >= target.SuccessThreshold {
			state.Status = models.StatusUp
			inc := state.CurrentIncident
			resolvedAt := res.Timestamp
			duration := resolvedAt.Sub(inc.StartedAt)
			inc.ResolvedAt = &resolvedAt
			inc.Duration = duration
			inc.Status = models.IncidentResolved

			if d.incidentRepo != nil {
				_ = d.incidentRepo.ResolveIncident(ctx, inc.ID, resolvedAt, duration)
			}

			alert = &models.AlertEvent{
				ID:         uuid.New().String(),
				IncidentID: inc.ID,
				TargetID:   target.ID,
				TargetName: target.Name,
				TargetURL:  target.URL,
				EventType:  models.AlertRecovery,
				Message:    fmt.Sprintf("Target %s RECOVERED after %s", target.Name, duration.Round(time.Second)),
				LatencyMs:  float64(res.Latency.Milliseconds()),
				StatusCode: res.StatusCode,
				Timestamp:  res.Timestamp,
			}

			state.CurrentIncident = nil
		}
	}

	if alert != nil && d.listener != nil {
		d.listener(alert)
	}

	return generatedIncident, alert
}

func (d *IncidentDetector) GetTargetStatus(targetID string) models.CheckStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if state, ok := d.states[targetID]; ok {
		return state.Status
	}
	return models.StatusUp
}
