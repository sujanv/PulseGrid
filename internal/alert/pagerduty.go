package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type PagerDutyChannel struct {
	routingKey string
	client     *http.Client
}

func NewPagerDutyChannel(routingKey string) *PagerDutyChannel {
	return &PagerDutyChannel{
		routingKey: routingKey,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *PagerDutyChannel) Type() models.ChannelType {
	return "PAGERDUTY"
}

func (p *PagerDutyChannel) Send(ctx context.Context, event *models.AlertEvent) error {
	action := "trigger"
	severity := "error"
	if event.EventType == models.AlertRecovery {
		action = "resolve"
		severity = "info"
	}

	dedupKey := fmt.Sprintf("pulsegrid-%s", event.TargetID)
	if event.IncidentID != "" {
		dedupKey = fmt.Sprintf("pulsegrid-inc-%s", event.IncidentID)
	}

	payload := map[string]interface{}{
		"routing_key":  p.routingKey,
		"event_action": action,
		"dedup_key":    dedupKey,
		"payload": map[string]interface{}{
			"summary":   event.Message,
			"source":    "PulseGrid",
			"severity":  severity,
			"timestamp": event.Timestamp.Format(time.RFC3339),
			"custom_details": map[string]interface{}{
				"target_name": event.TargetName,
				"target_url":  event.TargetURL,
				"latency_ms":  event.LatencyMs,
				"status_code": event.StatusCode,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://events.pagerduty.com/v2/enqueue", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty events api returned %d", resp.StatusCode)
	}
	return nil
}
