package network

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type GeoDNSRegionPolicy struct {
	Region          string   `json:"region"`
	ExpectedPrefixes []string `json:"expected_prefixes"` // e.g. "198.51.", "192.0.2."
}

type GeoDNSCheckResult struct {
	Region         string        `json:"region"`
	ResolvedIPs    []string      `json:"resolved_ips"`
	MatchesPolicy  bool          `json:"matches_policy"`
	Latency        time.Duration `json:"latency"`
	ErrorMessage   string        `json:"error_message,omitempty"`
}

type GeoDNSSummary struct {
	Domain       string               `json:"domain"`
	Healthy      bool                 `json:"healthy"`
	Results      []GeoDNSCheckResult  `json:"results"`
	EvaluatedAt  time.Time            `json:"evaluated_at"`
}

type GeoDNSResolver struct {
	policies map[string][]GeoDNSRegionPolicy // domain -> policies
}

func NewGeoDNSResolver() *GeoDNSResolver {
	return &GeoDNSResolver{
		policies: make(map[string][]GeoDNSRegionPolicy),
	}
}

func (r *GeoDNSResolver) RegisterPolicy(domain string, policies []GeoDNSRegionPolicy) {
	r.policies[domain] = policies
}

func (r *GeoDNSResolver) VerifyGeoSteering(ctx context.Context, domain string, simulatedResolutions map[string][]string) GeoDNSSummary {
	policies, exists := r.policies[domain]
	summary := GeoDNSSummary{
		Domain:      domain,
		Healthy:     true,
		Results:     make([]GeoDNSCheckResult, 0),
		EvaluatedAt: time.Now().UTC(),
	}

	if !exists {
		return summary
	}

	for _, policy := range policies {
		resolvedIPs, ok := simulatedResolutions[policy.Region]
		if !ok || len(resolvedIPs) == 0 {
			summary.Healthy = false
			summary.Results = append(summary.Results, GeoDNSCheckResult{
				Region:        policy.Region,
				MatchesPolicy: false,
				ErrorMessage:  fmt.Sprintf("no DNS resolution returned for region %s", policy.Region),
			})
			continue
		}

		matches := false
		for _, ip := range resolvedIPs {
			for _, prefix := range policy.ExpectedPrefixes {
				if strings.HasPrefix(ip, prefix) {
					matches = true
					break
				}
			}
			if matches {
				break
			}
		}

		if !matches {
			summary.Healthy = false
		}

		summary.Results = append(summary.Results, GeoDNSCheckResult{
			Region:        policy.Region,
			ResolvedIPs:   resolvedIPs,
			MatchesPolicy: matches,
			Latency:       18 * time.Millisecond,
		})
	}

	return summary
}
