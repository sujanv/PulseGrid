package integrations

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type TeamsCardFormatter struct{}

func NewTeamsCardFormatter() *TeamsCardFormatter {
	return &TeamsCardFormatter{}
}

func (f *TeamsCardFormatter) FormatCard(event *models.AlertEvent) ([]byte, error) {
	statusColor := "attention" // Red
	if event.EventType == "RESOLVED" {
		statusColor = "good" // Green
	}

	card := map[string]interface{}{
		"type":    "message",
		"summary": fmt.Sprintf("PulseGrid Alert: %s is %s", event.TargetName, event.EventType),
		"attachments": []map[string]interface{}{
			{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"content": map[string]interface{}{
					"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
					"type":    "AdaptiveCard",
					"version": "1.4",
					"body": []map[string]interface{}{
						{
							"type":   "TextBlock",
							"size":   "Medium",
							"weight": "Bolder",
							"color":  statusColor,
							"text":   fmt.Sprintf("⚡ PulseGrid %s: %s", event.EventType, event.TargetName),
						},
						{
							"type": "FactSet",
							"facts": []map[string]string{
								{"title": "Target URL:", "value": event.TargetURL},
								{"title": "Timestamp:", "value": event.Timestamp.Format(time.RFC3339)},
								{"title": "Error Message:", "value": event.Message},
							},
						},
					},
					"actions": []map[string]interface{}{
						{
							"type":  "Action.OpenUrl",
							"title": "View in PulseGrid",
							"url":   fmt.Sprintf("https://pulsegrid.io/targets/%s", event.TargetID),
						},
					},
				},
			},
		},
	}

	return json.Marshal(card)
}

type DiscordEmbedFormatter struct{}

func NewDiscordEmbedFormatter() *DiscordEmbedFormatter {
	return &DiscordEmbedFormatter{}
}

func (d *DiscordEmbedFormatter) FormatEmbed(event *models.AlertEvent) ([]byte, error) {
	color := 15158332 // Red
	if event.EventType == "RESOLVED" {
		color = 3066993 // Green
	}

	payload := map[string]interface{}{
		"username":   "PulseGrid Monitor",
		"avatar_url": "https://pulsegrid.io/logo.png",
		"embeds": []map[string]interface{}{
			{
				"title":       fmt.Sprintf("⚡ Alert: %s is %s", event.TargetName, event.EventType),
				"description": event.Message,
				"url":         fmt.Sprintf("https://pulsegrid.io/targets/%s", event.TargetID),
				"color":       color,
				"fields": []map[string]interface{}{
					{"name": "Target URL", "value": fmt.Sprintf("`%s`", event.TargetURL), "inline": false},
					{"name": "Timestamp", "value": event.Timestamp.Format(time.RFC3339), "inline": true},
				},
				"footer": map[string]string{
					"text": "PulseGrid Enterprise Monitoring",
				},
			},
		},
	}

	return json.Marshal(payload)
}
