package remediation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type ActionType string

const (
	ActionWebhook ActionType = "WEBHOOK"
	ActionScript  ActionType = "SCRIPT"
)

type RemediationRule struct {
	ID        string        `json:"id"`
	TargetID  string        `json:"target_id"`
	Action    ActionType    `json:"action"`
	URL       string        `json:"url"`
	Cooldown  time.Duration `json:"cooldown"`
	Enabled   bool          `json:"enabled"`
}

type RemediationEngine struct {
	circuit *CircuitBreaker
	client  *http.Client
}

func NewRemediationEngine(cb *CircuitBreaker) *RemediationEngine {
	return &RemediationEngine{
		circuit: cb,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (e *RemediationEngine) Trigger(ctx context.Context, rule *RemediationRule, inc *models.Incident) error {
	if !rule.Enabled {
		return nil
	}

	if !e.circuit.Allow(rule.TargetID, rule.Cooldown) {
		return fmt.Errorf("remediation action in cooldown for target %s", rule.TargetID)
	}

	payload := map[string]interface{}{
		"action_id":   uuid.New().String(),
		"target_id":   rule.TargetID,
		"incident_id": inc.ID,
		"root_cause":  inc.RootCause,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", rule.URL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PulseGrid-Remediation-Engine/1.0")

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("remediation webhook returned %d", resp.StatusCode)
	}
	return nil
}
