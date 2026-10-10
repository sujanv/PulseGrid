package alert

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type EmailChannel struct {
	smtpHost  string
	smtpPort  int
	smtpUser  string
	smtpPass  string
	fromEmail string
	toEmail   string
}

func NewEmailChannel(host string, port int, user, pass, from, to string) *EmailChannel {
	return &EmailChannel{
		smtpHost:  host,
		smtpPort:  port,
		smtpUser:  user,
		smtpPass:  pass,
		fromEmail: from,
		toEmail:   to,
	}
}

func (e *EmailChannel) Type() models.ChannelType {
	return models.ChannelEmail
}

func (e *EmailChannel) Send(ctx context.Context, event *models.AlertEvent) error {
	if e.smtpHost == "" || e.toEmail == "" {
		return nil // skip if not configured
	}

	subject := fmt.Sprintf("[PulseGrid] %s: %s", event.EventType, event.TargetName)
	color := "#10b981"
	if event.EventType == models.AlertOutage {
		color = "#ef4444"
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background: #f3f4f6; padding: 20px;">
  <div style="max-width: 600px; margin: auto; background: #ffffff; border-radius: 8px; border-left: 6px solid %s; padding: 24px;">
    <h2 style="margin-top: 0; color: #111827;">PulseGrid Alert: %s</h2>
    <p style="font-size: 16px; color: #374151;">%s</p>
    <table style="width: 100%%; border-collapse: collapse; margin-top: 16px;">
      <tr><td style="padding: 8px 0; color: #6b7280;">Target:</td><td style="font-weight: bold;">%s</td></tr>
      <tr><td style="padding: 8px 0; color: #6b7280;">Endpoint:</td><td><code>%s</code></td></tr>
      <tr><td style="padding: 8px 0; color: #6b7280;">Timestamp:</td><td>%s</td></tr>
    </table>
    <div style="margin-top: 24px; padding-top: 16px; border-top: 1px solid #e5e7eb; font-size: 12px; color: #9ca3af;">
      PulseGrid Uptime Monitoring Service
    </div>
  </div>
</body>
</html>`, color, event.EventType, event.Message, event.TargetName, event.TargetURL, event.Timestamp.Format(time.RFC1123))

	headers := make(map[string]string)
	headers["From"] = e.fromEmail
	headers["To"] = e.toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s
", k, v))
	}
	msg.WriteString("
" + htmlBody)

	auth := smtp.PlainAuth("", e.smtpUser, e.smtpPass, e.smtpHost)
	addr := fmt.Sprintf("%s:%d", e.smtpHost, e.smtpPort)

	return smtp.SendMail(addr, auth, e.fromEmail, []string{e.toEmail}, []byte(msg.String()))
}
