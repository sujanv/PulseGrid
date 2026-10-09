package models

import (
	"time"
)

type CheckStatus string

const (
	StatusUp       CheckStatus = "UP"
	StatusDown     CheckStatus = "DOWN"
	StatusDegraded CheckStatus = "DEGRADED"
)

type LatencyBreakdown struct {
	DNSLookup       time.Duration `json:"dns_lookup_ms"`
	TCPConnect      time.Duration `json:"tcp_connect_ms"`
	TLSHandshake    time.Duration `json:"tls_handshake_ms"`
	TTFB            time.Duration `json:"ttfb_ms"`
	ContentTransfer time.Duration `json:"content_transfer_ms"`
	Total           time.Duration `json:"total_ms"`
}

type CheckResult struct {
	ID                  string           `json:"id"`
	TargetID            string           `json:"target_id"`
	Timestamp           time.Time        `json:"timestamp"`
	Status              CheckStatus      `json:"status"`
	StatusCode          int              `json:"status_code,omitempty"`
	ResponseBytes       int64            `json:"response_bytes"`
	Latency             time.Duration    `json:"latency"`
	Breakdown           LatencyBreakdown `json:"breakdown"`
	ErrorMessage        string           `json:"error_message,omitempty"`
	CertificateDaysLeft int              `json:"cert_days_left,omitempty"`
	ResolvedIPs         []string         `json:"resolved_ips,omitempty"`
}

func (r *CheckResult) IsHealthy() bool {
	return r.Status == StatusUp
}
