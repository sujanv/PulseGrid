package alert

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type WebhookChannel struct {
	url       string
	secretKey string
	client    *http.Client
}

func NewWebhookChannel(url, secretKey string) *WebhookChannel {
	return &WebhookChannel{
		url:       url,
		secretKey: secretKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (w *WebhookChannel) Type() models.ChannelType {
	return models.ChannelWebhook
}

func (w *WebhookChannel) Send(ctx context.Context, event *models.AlertEvent) error {
	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", w.url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "PulseGrid-Alert-Webhook/1.0")

	if w.secretKey != "" {
		mac := hmac.New(sha256.New, []byte(w.secretKey))
		mac.Write(payloadBytes)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-PulseGrid-Signature", fmt.Sprintf("sha256=%s", signature))
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook responded with status: %d", resp.StatusCode)
	}

	return nil
}
