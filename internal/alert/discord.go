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

type DiscordChannel struct {
	webhookURL string
	client     *http.Client
}

func NewDiscordChannel(webhookURL string) *DiscordChannel {
	return &DiscordChannel{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (d *DiscordChannel) Type() models.ChannelType {
	return models.ChannelDiscord
}

func (d *DiscordChannel) Send(ctx context.Context, event *models.AlertEvent) error {
	color := 0x2ECC71 // Green
	if event.EventType == models.AlertOutage {
		color = 0xE74C3C // Red
	}

	payload := map[string]interface{}{
		"username": "PulseGrid Alert",
		"embeds": []map[string]interface{}{
			{
				"title":       fmt.Sprintf("PulseGrid Alert: %s", event.EventType),
				"description": event.Message,
				"color":       color,
				"fields": []map[string]interface{}{
					{"name": "Target", "value": event.TargetName, "inline": true},
					{"name": "URL", "value": event.TargetURL, "inline": true},
					{"name": "Latency", "value": fmt.Sprintf("%.2f ms", event.LatencyMs), "inline": true},
				},
				"timestamp": event.Timestamp.Format(time.RFC3339),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", d.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}
