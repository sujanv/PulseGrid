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

type TelegramChannel struct {
	botToken string
	chatID   string
	client   *http.Client
}

func NewTelegramChannel(botToken, chatID string) *TelegramChannel {
	return &TelegramChannel{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *TelegramChannel) Type() models.ChannelType {
	return "TELEGRAM"
}

func (t *TelegramChannel) Send(ctx context.Context, event *models.AlertEvent) error {
	emoji := "🚨"
	statusText := "OUTAGE DETECTED"
	if event.EventType == models.AlertRecovery {
		emoji = "✅"
		statusText = "SERVICE RECOVERED"
	}

	text := fmt.Sprintf("%s *PulseGrid: %s*\n\n*Target:* %s\n*Endpoint:* %s\n*Message:* %s\n*Timestamp:* %s",
		emoji, statusText, event.TargetName, event.TargetURL, event.Message, event.Timestamp.Format(time.RFC1123))

	payload := map[string]interface{}{
		"chat_id":    t.chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(body))
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
		return fmt.Errorf("telegram api error: status %d", resp.StatusCode)
	}
	return nil
}
