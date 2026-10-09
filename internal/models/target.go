package models

import (
	"errors"
	"net/url"
	"time"
)

type TargetType string

const (
	TargetHTTP  TargetType = "HTTP"
	TargetHTTPS TargetType = "HTTPS"
	TargetTCP   TargetType = "TCP"
	TargetDNS   TargetType = "DNS"
	TargetTLS   TargetType = "TLS"
)

type HTTPMethod string

const (
	MethodGET  HTTPMethod = "GET"
	MethodPOST HTTPMethod = "POST"
	MethodHEAD HTTPMethod = "HEAD"
	MethodPUT  HTTPMethod = "PUT"
)

type Target struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	URL                 string            `json:"url"`
	Host                string            `json:"host,omitempty"`
	Port                int               `json:"port,omitempty"`
	Type                TargetType        `json:"type"`
	Interval            time.Duration     `json:"interval"`
	Timeout             time.Duration     `json:"timeout"`
	Method              HTTPMethod        `json:"method,omitempty"`
	Headers             map[string]string `json:"headers,omitempty"`
	Body                string            `json:"body,omitempty"`
	ExpectedStatusCodes []int             `json:"expected_status_codes,omitempty"`
	ExpectedKeyword     string            `json:"expected_keyword,omitempty"`
	FailureThreshold    int               `json:"failure_threshold"`
	SuccessThreshold    int               `json:"success_threshold"`
	Active              bool              `json:"active"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

type TargetSummary struct {
	Target           *Target      `json:"target"`
	LastResult       *CheckResult `json:"last_result,omitempty"`
	CurrentState     CheckStatus  `json:"current_state"`
	UptimePercentage float64      `json:"uptime_percentage"`
	AvgLatencyMs     float64      `json:"avg_latency_ms"`
}

func (t *Target) Validate() error {
	if t.Name == "" {
		return errors.New("target name is required")
	}
	if t.Interval < 1*time.Second {
		t.Interval = 30 * time.Second
	}
	if t.Timeout < 100*time.Millisecond {
		t.Timeout = 10 * time.Second
	}
	if t.FailureThreshold < 1 {
		t.FailureThreshold = 3
	}
	if t.SuccessThreshold < 1 {
		t.SuccessThreshold = 2
	}

	switch t.Type {
	case TargetHTTP, TargetHTTPS:
		u, err := url.ParseRequestURI(t.URL)
		if err != nil {
			return errors.New("invalid target URL")
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return errors.New("URL scheme must be http or https")
		}
		if len(t.ExpectedStatusCodes) == 0 {
			t.ExpectedStatusCodes = []int{200, 201, 204}
		}
		if t.Method == "" {
			t.Method = MethodGET
		}
	case TargetTCP, TargetTLS:
		if t.Host == "" && t.URL != "" {
			u, err := url.Parse(t.URL)
			if err == nil && u.Hostname() != "" {
				t.Host = u.Hostname()
			}
		}
		if t.Host == "" {
			return errors.New("target host is required for TCP/TLS")
		}
		if t.Port <= 0 || t.Port > 65535 {
			if t.Type == TargetTLS {
				t.Port = 443
			} else {
				return errors.New("valid port (1-65535) is required for TCP")
			}
		}
	case TargetDNS:
		if t.Host == "" && t.URL != "" {
			u, err := url.Parse(t.URL)
			if err == nil && u.Hostname() != "" {
				t.Host = u.Hostname()
			}
		}
		if t.Host == "" {
			return errors.New("target domain/host is required for DNS check")
		}
	default:
		t.Type = TargetHTTP
	}

	return nil
}
