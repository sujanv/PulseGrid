package rca

import (
	"fmt"
	"strings"
	"time"
)

type IncidentTimelineEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Source      string    `json:"source"`
}

type IncidentData struct {
	IncidentID      string                  `json:"incident_id"`
	TargetID        string                  `json:"target_id"`
	TargetName      string                  `json:"target_name"`
	TargetURL       string                  `json:"target_url"`
	RootCause       RootCauseType           `json:"root_cause"`
	Recommendation  string                  `json:"recommendation"`
	StartedAt       time.Time               `json:"started_at"`
	ResolvedAt      time.Time               `json:"resolved_at"`
	Timeline        []IncidentTimelineEntry `json:"timeline"`
	AffectedRegions []string                `json:"affected_regions"`
}

type PostMortemGenerator struct{}

func NewPostMortemGenerator() *PostMortemGenerator {
	return &PostMortemGenerator{}
}

func (g *PostMortemGenerator) GenerateMarkdown(data *IncidentData) string {
	duration := data.ResolvedAt.Sub(data.StartedAt)
	if duration < 0 {
		duration = 0
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# 📋 SRE Incident Post-Mortem: %s\n\n", data.IncidentID))
	sb.WriteString(fmt.Sprintf("**Target:** `%s` (%s)\n", data.TargetName, data.TargetURL))
	sb.WriteString(fmt.Sprintf("**Outage Duration:** %v\n", duration.Round(time.Second)))
	sb.WriteString(fmt.Sprintf("**Root Cause Classification:** `%s`\n", data.RootCause))
	sb.WriteString(fmt.Sprintf("**Primary Recommendation:** %s\n\n", data.Recommendation))

	sb.WriteString("## 🌍 Affected Geographic Regions\n\n")
	if len(data.AffectedRegions) > 0 {
		for _, reg := range data.AffectedRegions {
			sb.WriteString(fmt.Sprintf("- %s\n", reg))
		}
	} else {
		sb.WriteString("- Global Outage (All Vantage Points)\n")
	}
	sb.WriteString("\n")

	sb.WriteString("## ⏱ Chronological Incident Timeline\n\n")
	sb.WriteString("| UTC Timestamp | Event Description | Source | Severity |\n")
	sb.WriteString("|---|---|---|---|\n")

	for _, entry := range data.Timeline {
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n",
			entry.Timestamp.Format("2006-01-02 15:04:05"),
			entry.Description,
			entry.Source,
			entry.Severity,
		))
	}

	sb.WriteString("\n## 🛠 Action Items & Preventative Measures\n\n")
	sb.WriteString("- [ ] Automated canary deployments with instant rollback triggers\n")
	sb.WriteString("- [ ] Fine-tune alert escalation thresholds and quiet hours\n")
	sb.WriteString("- [ ] Verify multi-region DNS failover health checks\n")

	return sb.String()
}
