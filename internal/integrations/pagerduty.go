package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type PagerDutyAction string

const (
	PDActionTrigger     PagerDutyAction = "trigger"
	PDActionAcknowledge PagerDutyAction = "acknowledge"
	PDActionResolve     PagerDutyAction = "resolve"
)

type PagerDutyPayload struct {
	Summary   string            `json:"summary"`
	Severity  string            `json:"severity"` // critical, error, warning, info
	Source    string            `json:"source"`
	Timestamp string            `json:"timestamp"`
	Custom    map[string]string `json:"custom_details,omitempty"`
}

type PagerDutyEvent struct {
	RoutingKey  string           `json:"routing_key"`
	EventAction PagerDutyAction  `json:"event_action"`
	DedupKey    string           `json:"dedup_key"`
	Payload     PagerDutyPayload `json:"payload"`
}

type PagerDutyClient struct {
	routingKey string
	endpoint   string
	httpClient *http.Client
}

func NewPagerDutyClient(routingKey string) *PagerDutyClient {
	return &PagerDutyClient{
		routingKey: routingKey,
		endpoint:   "https://events.pagerduty.com/v2/enqueue",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *PagerDutyClient) Dispatch(ctx context.Context, action PagerDutyAction, target *models.Target, errMsg string) error {
	if c.routingKey == "" {
		return nil // skip if unconfigured
	}

	dedupKey := fmt.Sprintf("pulsegrid-target-%s", target.ID)
	summary := fmt.Sprintf("PulseGrid Alert: Target '%s' is DOWN", target.Name)
	if action == PDActionResolve {
		summary = fmt.Sprintf("PulseGrid Alert: Target '%s' has RECOVERED", target.Name)
	}

	event := PagerDutyEvent{
		RoutingKey:  c.routingKey,
		EventAction: action,
		DedupKey:    dedupKey,
		Payload: PagerDutyPayload{
			Summary:   summary,
			Severity:  "critical",
			Source:    target.URL,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Custom: map[string]string{
				"target_id": target.ID,
				"url":       target.URL,
				"error":     errMsg,
			},
		},
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty api error: status %s", resp.Status)
	}

	return nil
}
