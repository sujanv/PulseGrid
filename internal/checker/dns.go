package checker

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type DNSChecker struct {
	resolver *net.Resolver
}

func NewDNSChecker() *DNSChecker {
	return &DNSChecker{
		resolver: net.DefaultResolver,
	}
}

func (c *DNSChecker) Check(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	host := target.Host
	if host == "" {
		host = target.URL
	}

	startTime := time.Now()
	ips, err := c.resolver.LookupHost(ctx, host)
	latency := time.Since(startTime)

	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      latency,
			ErrorMessage: fmt.Sprintf("DNS lookup failed for %s: %v", host, err),
			Breakdown: models.LatencyBreakdown{
				DNSLookup: latency,
				Total:     latency,
			},
		}, nil
	}

	if len(ips) == 0 {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      latency,
			ErrorMessage: fmt.Sprintf("DNS lookup returned 0 IP records for %s", host),
		}, nil
	}

	return &models.CheckResult{
		ID:          uuid.New().String(),
		TargetID:    target.ID,
		Timestamp:   time.Now().UTC(),
		Status:      models.StatusUp,
		Latency:     latency,
		ResolvedIPs: ips,
		Breakdown: models.LatencyBreakdown{
			DNSLookup: latency,
			Total:     latency,
		},
	}, nil
}
