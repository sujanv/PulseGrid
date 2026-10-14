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

type SlackChannel struct {
	webhookURL string
	client     *http.Client
}

func NewSlackChannel(webhookURL string) *SlackChannel {
	return &SlackChannel{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *SlackChannel) Type() models.ChannelType {
	return models.ChannelSlack
}

func (s *SlackChannel) Send(ctx context.Context, event *models.AlertEvent) error {
	color := "#36a64f" // Green
	emoji := ":white_check_mark:"
	if event.EventType == models.AlertOutage {
		color = "#ff0000" // Red
		emoji = ":rotating_light:"
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"blocks": []map[string]interface{}{
					{
						"type": "header",
						"text": map[string]string{
							"type": "plain_text",
							"text": fmt.Sprintf("%s PulseGrid Alert: %s", emoji, event.EventType),
						},
					},
					{
						"type": "section",
						"fields": []map[string]string{
							{"type": "mrkdwn", "text": fmt.Sprintf("*Target:*\n%s", event.TargetName)},
							{"type": "mrkdwn", "text": fmt.Sprintf("*Endpoint:*\n`%s`", event.TargetURL)},
							{"type": "mrkdwn", "text": fmt.Sprintf("*Status:*\n%s", event.EventType)},
							{"type": "mrkdwn", "text": fmt.Sprintf("*Time:*\n%s", event.Timestamp.Format(time.RFC3339))},
						},
					},
					{
						"type": "section",
						"text": map[string]string{
							"type": "mrkdwn",
							"text": fmt.Sprintf("*Details:*\n%s", event.Message),
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}
