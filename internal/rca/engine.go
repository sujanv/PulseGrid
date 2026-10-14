package rca

import (
	"strings"
	"time"
)

type RootCauseType string

const (
	CauseDNSFailure        RootCauseType = "DNS_RESOLUTION_FAILURE"
	CauseTLSTimeout        RootCauseType = "TLS_HANDSHAKE_TIMEOUT"
	CauseCertExpired       RootCauseType = "SSL_CERTIFICATE_EXPIRED"
	CauseGateway502        RootCauseType = "UPSTREAM_BAD_GATEWAY_502"
	CauseGateway504        RootCauseType = "UPSTREAM_GATEWAY_TIMEOUT_504"
	CauseNetworkTimeout    RootCauseType = "NETWORK_SOCKET_TIMEOUT"
	CauseConnectionRefused RootCauseType = "CONNECTION_REFUSED_PORT_DOWN"
	CauseUnknown           RootCauseType = "UNKNOWN_ERROR"
)

type DiagnosticReport struct {
	TargetID         string        `json:"target_id"`
	RootCause        RootCauseType `json:"root_cause"`
	ConfidenceScore  float64       `json:"confidence_score"`
	Recommendation   string        `json:"recommendation"`
	FailedAt         time.Time     `json:"failed_at"`
}

type CorrelatedCluster struct {
	ClusterID        string        `json:"cluster_id"`
	CommonFactor     string        `json:"common_factor"` // e.g. "CloudFlare DNS", "AWS us-east-1"
	AffectedTargets  []string      `json:"affected_targets"`
	SharedCause      RootCauseType `json:"shared_cause"`
	DetectedAt       time.Time     `json:"detected_at"`
}

type RCAEngine struct{}

func NewRCAEngine() *RCAEngine {
	return &RCAEngine{}
}

func (e *RCAEngine) Diagnose(targetID, errMessage string, statusCode int) DiagnosticReport {
	lower := strings.ToLower(errMessage)
	cause := CauseUnknown
	rec := "Inspect application logs and network routing."
	conf := 0.70

	switch {
	case strings.Contains(lower, "no such host") || strings.Contains(lower, "dns") || strings.Contains(lower, "lookup"):
		cause = CauseDNSFailure
		rec = "Check authoritative nameserver status and domain DNS registrar delegation."
		conf = 0.95
	case strings.Contains(lower, "certificate has expired"):
		cause = CauseCertExpired
		rec = "Renew SSL/TLS certificate immediately with Let's Encrypt or your certificate authority."
		conf = 0.99
	case strings.Contains(lower, "tls handshake timeout"):
		cause = CauseTLSTimeout
		rec = "Investigate intermediate firewall MTU sizing or reverse proxy TLS listener overload."
		conf = 0.90
	case statusCode == 502:
		cause = CauseGateway502
		rec = "Backend application pod or upstream service crashed or is restarting."
		conf = 0.92
	case statusCode == 504:
		cause = CauseGateway504
		rec = "Upstream database or dependency slow query causing reverse proxy gateway timeout."
		conf = 0.91
	case strings.Contains(lower, "connection refused"):
		cause = CauseConnectionRefused
		rec = "Service process is stopped or listening on incorrect network port."
		conf = 0.95
	case strings.Contains(lower, "timeout") || strings.Contains(lower, "deadline exceeded"):
		cause = CauseNetworkTimeout
		rec = "High latency or packet drop on network route; increase probe timeout if expected."
		conf = 0.85
	}

	return DiagnosticReport{
		TargetID:        targetID,
		RootCause:       cause,
		ConfidenceScore: conf,
		Recommendation:  rec,
		FailedAt:        time.Now().UTC(),
	}
}

func (e *RCAEngine) FindClusters(reports []DiagnosticReport) []CorrelatedCluster {
	grouped := make(map[RootCauseType][]string)
	for _, r := range reports {
		grouped[r.RootCause] = append(grouped[r.RootCause], r.TargetID)
	}

	var clusters []CorrelatedCluster
	for cause, targets := range grouped {
		if len(targets) >= 2 {
			clusters = append(clusters, CorrelatedCluster{
				ClusterID:       string(cause) + "-cluster",
				CommonFactor:    string(cause),
				AffectedTargets: targets,
				SharedCause:     cause,
				DetectedAt:      time.Now().UTC(),
			})
		}
	}
	return clusters
}
