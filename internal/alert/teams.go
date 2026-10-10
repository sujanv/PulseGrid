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

type TeamsChannel struct {
	webhookURL string
	client     *http.Client
}

func NewTeamsChannel(webhookURL string) *TeamsChannel {
	return &TeamsChannel{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *TeamsChannel) Type() models.ChannelType {
	return "TEAMS"
}

func (t *TeamsChannel) Send(ctx context.Context, event *models.AlertEvent) error {
	themeColor := "E81123" // Red
	if event.EventType == models.AlertRecovery {
		themeColor = "107C41" // Green
	}

	payload := map[string]interface{}{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": themeColor,
		"summary":    event.Message,
		"sections": []map[string]interface{}{
			{
				"activityTitle": fmt.Sprintf("PulseGrid Alert: %s", event.EventType),
				"activitySubtitle": event.Timestamp.Format(time.RFC1123),
				"facts": []map[string]string{
					{"name": "Target", "value": event.TargetName},
					{"name": "Endpoint", "value": event.TargetURL},
					{"name": "Latency", "value": fmt.Sprintf("%.2f ms", event.LatencyMs)},
					{"name": "Status Code", "value": fmt.Sprintf("%d", event.StatusCode)},
					{"name": "Details", "value": event.Message},
				},
				"markdown": true,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("teams webhook responded with %d", resp.StatusCode)
	}
	return nil
}
