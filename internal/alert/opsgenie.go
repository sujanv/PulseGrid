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

type OpsGenieChannel struct {
	apiKey string
	client *http.Client
}

func NewOpsGenieChannel(apiKey string) *OpsGenieChannel {
	return &OpsGenieChannel{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (o *OpsGenieChannel) Type() models.ChannelType {
	return "OPSGENIE"
}

func (o *OpsGenieChannel) Send(ctx context.Context, event *models.AlertEvent) error {
	apiURL := "https://api.opsgenie.com/v2/alerts"
	alias := fmt.Sprintf("pulsegrid-%s", event.TargetID)

	var payload map[string]interface{}
	if event.EventType == models.AlertRecovery {
		apiURL = fmt.Sprintf("https://api.opsgenie.com/v2/alerts/%s/close?identifierType=alias", alias)
		payload = map[string]interface{}{
			"source": "PulseGrid",
			"note":   "Service has recovered",
		}
	} else {
		payload = map[string]interface{}{
			"message":  event.Message,
			"alias":    alias,
			"source":   "PulseGrid Uptime Monitor",
			"priority": "P2",
			"details": map[string]string{
				"target_id":   event.TargetID,
				"target_name": event.TargetName,
				"endpoint":    event.TargetURL,
			},
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("GenieKey %s", o.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opsgenie returned status: %d", resp.StatusCode)
	}
	return nil
}
