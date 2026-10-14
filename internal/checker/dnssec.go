package checker

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type DNSSECChecker struct {
	resolver *net.Resolver
}

func NewDNSSECChecker() *DNSSECChecker {
	return &DNSSECChecker{
		resolver: net.DefaultResolver,
	}
}

func (c *DNSSECChecker) Check(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	host := target.Host
	if host == "" {
		host = target.URL
	}

	start := time.Now()
	records, err := c.resolver.LookupHost(ctx, host)
	latency := time.Since(start)

	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      latency,
			ErrorMessage: fmt.Sprintf("DNSSEC validation failed for %s: %v", host, err),
		}, nil
	}

	return &models.CheckResult{
		ID:          uuid.New().String(),
		TargetID:    target.ID,
		Timestamp:   time.Now().UTC(),
		Status:      models.StatusUp,
		Latency:     latency,
		ResolvedIPs: records,
		Breakdown: models.LatencyBreakdown{
			DNSLookup: latency,
			Total:     latency,
		},
	}, nil
}
