package models

import (
	"time"
)

type IncidentStatus string

const (
	IncidentOpen          IncidentStatus = "OPEN"
	IncidentInvestigating IncidentStatus = "INVESTIGATING"
	IncidentResolved      IncidentStatus = "RESOLVED"
)

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityWarning  Severity = "WARNING"
)

type Incident struct {
	ID                  string         `json:"id"`
	TargetID            string         `json:"target_id"`
	TargetName          string         `json:"target_name"`
	StartedAt           time.Time      `json:"started_at"`
	ResolvedAt          *time.Time     `json:"resolved_at,omitempty"`
	Duration            time.Duration  `json:"duration,omitempty"`
	Status              IncidentStatus `json:"status"`
	RootCause           string         `json:"root_cause"`
	ConsecutiveFailures int            `json:"consecutive_failures"`
	Severity            Severity       `json:"severity"`
}
