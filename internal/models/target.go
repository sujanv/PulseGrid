package models

import (
	"errors"
	"net/url"
	"time"
)

type TargetType string

const (
	TargetHTTP      TargetType = "HTTP"
	TargetHTTPS     TargetType = "HTTPS"
	TargetTCP       TargetType = "TCP"
	TargetDNS       TargetType = "DNS"
	TargetTLS       TargetType = "TLS"
	TargetPing      TargetType = "PING"
	TargetWebSocket TargetType = "WEBSOCKET"
	TargetGRPC      TargetType = "GRPC"
	TargetSynthetic TargetType = "SYNTHETIC"
)

type HTTPMethod string

const (
	MethodGET  HTTPMethod = "GET"
	MethodPOST HTTPMethod = "POST"
	MethodHEAD HTTPMethod = "HEAD"
	MethodPUT  HTTPMethod = "PUT"
)

type AssertionRule struct {
	Type     string `json:"type"`     // JSONPATH, REGEX, HEADER
	Property string `json:"property"` // Path, Header name, etc.
	Operator string `json:"operator"` // equals, contains, greater_than, less_than
	Expected string `json:"expected"`
}

type Target struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	URL                 string            `json:"url"`
	Host                string            `json:"host,omitempty"`
	Port                int               `json:"port,omitempty"`
	Type                TargetType        `json:"type"`
	Interval            time.Duration     `json:"interval"`
	Timeout             time.Duration     `json:"timeout"`
	DegradedThresholdMs int               `json:"degraded_threshold_ms,omitempty"`
	Method              HTTPMethod        `json:"method,omitempty"`
	Headers             map[string]string `json:"headers,omitempty"`
	Body                string            `json:"body,omitempty"`
	ExpectedStatusCodes []int             `json:"expected_status_codes,omitempty"`
	ExpectedKeyword     string            `json:"expected_keyword,omitempty"`
	Assertions          []AssertionRule   `json:"assertions,omitempty"`
	WorkflowJSON        string            `json:"workflow_json,omitempty"`
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
	case TargetHTTP, TargetHTTPS, TargetWebSocket:
		u, err := url.ParseRequestURI(t.URL)
		if err != nil {
			return errors.New("invalid target URL")
		}
		if t.Type == TargetWebSocket {
			if u.Scheme != "ws" && u.Scheme != "wss" && u.Scheme != "http" && u.Scheme != "https" {
				return errors.New("WebSocket scheme must be ws, wss, http, or https")
			}
		} else {
			if u.Scheme != "http" && u.Scheme != "https" {
				return errors.New("URL scheme must be http or https")
			}
		}
		if len(t.ExpectedStatusCodes) == 0 && t.Type != TargetWebSocket {
			t.ExpectedStatusCodes = []int{200, 201, 204}
		}
		if t.Method == "" {
			t.Method = MethodGET
		}
	case TargetTCP, TargetTLS, TargetPing, TargetGRPC:
		if t.Host == "" && t.URL != "" {
			u, err := url.Parse(t.URL)
			if err == nil && u.Hostname() != "" {
				t.Host = u.Hostname()
			}
		}
		if t.Host == "" {
			return errors.New("target host is required")
		}
		if t.Port <= 0 || t.Port > 65535 {
			if t.Type == TargetTLS {
				t.Port = 443
			} else if t.Type == TargetGRPC {
				t.Port = 50051
			} else if t.Type == TargetTCP {
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
	case TargetSynthetic:
		if t.WorkflowJSON == "" {
			return errors.New("workflow configuration JSON is required for synthetic target")
		}
	default:
		t.Type = TargetHTTP
	}

	return nil
}
