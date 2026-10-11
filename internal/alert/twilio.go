package alert

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type TwilioChannel struct {
	accountSID string
	authToken  string
	fromPhone  string
	toPhone    string
	client     *http.Client
}

func NewTwilioChannel(sid, token, from, to string) *TwilioChannel {
	return &TwilioChannel{
		accountSID: sid,
		authToken:  token,
		fromPhone:  from,
		toPhone:    to,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *TwilioChannel) Type() models.ChannelType {
	return "TWILIO_SMS"
}

func (t *TwilioChannel) Send(ctx context.Context, event *models.AlertEvent) error {
	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSID)

	msgText := fmt.Sprintf("[PulseGrid] %s: %s is %s", event.EventType, event.TargetName, event.EventType)
	if len(msgText) > 160 {
		msgText = msgText[:157] + "..."
	}

	data := url.Values{}
	data.Set("From", t.fromPhone)
	data.Set("To", t.toPhone)
	data.Set("Body", msgText)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(t.accountSID, t.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("twilio api returned %d", resp.StatusCode)
	}
	return nil
}
