package models

import (
	"time"
)

type ChannelType string

const (
	ChannelWebhook ChannelType = "WEBHOOK"
	ChannelSlack   ChannelType = "SLACK"
	ChannelDiscord ChannelType = "DISCORD"
	ChannelEmail   ChannelType = "EMAIL"
)

type AlertChannel struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Type        ChannelType `json:"type"`
	Destination string      `json:"destination"` // URL or Email address
	SecretKey   string      `json:"secret_key,omitempty"`
	Enabled     bool        `json:"enabled"`
	CreatedAt   time.Time   `json:"created_at"`
}

type AlertEventType string

const (
	AlertOutage    AlertEventType = "OUTAGE"
	AlertRecovery  AlertEventType = "RECOVERY"
	AlertSSLExpiry AlertEventType = "SSL_EXPIRY"
	AlertFlapping  AlertEventType = "FLAPPING"
)

type AlertEvent struct {
	ID         string         `json:"id"`
	IncidentID string         `json:"incident_id,omitempty"`
	TargetID   string         `json:"target_id"`
	TargetName string         `json:"target_name"`
	TargetURL  string         `json:"target_url"`
	EventType  AlertEventType `json:"event_type"`
	Message    string         `json:"message"`
	LatencyMs  float64        `json:"latency_ms,omitempty"`
	StatusCode int            `json:"status_code,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

type AlertLog struct {
	ID           string    `json:"id"`
	EventID      string    `json:"event_id"`
	ChannelID    string    `json:"channel_id"`
	ChannelType  string    `json:"channel_type"`
	Status       string    `json:"status"` // SENT, FAILED
	ErrorMessage string    `json:"error_message,omitempty"`
	SentAt       time.Time `json:"sent_at"`
}
