package checker

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type TCPChecker struct{}

func NewTCPChecker() *TCPChecker {
	return &TCPChecker{}
}

func (c *TCPChecker) Check(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	address := fmt.Sprintf("%s:%d", target.Host, target.Port)
	dialer := &net.Dialer{}

	startTime := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", address)
	latency := time.Since(startTime)

	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      latency,
			ErrorMessage: fmt.Sprintf("TCP dial error to %s: %v", address, err),
		}, nil
	}
	defer conn.Close()

	var resolvedIPs []string
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if remoteAddr := tcpConn.RemoteAddr(); remoteAddr != nil {
			resolvedIPs = append(resolvedIPs, remoteAddr.String())
		}
	}

	return &models.CheckResult{
		ID:            uuid.New().String(),
		TargetID:      target.ID,
		Timestamp:     time.Now().UTC(),
		Status:        models.StatusUp,
		Latency:       latency,
		ResolvedIPs:   resolvedIPs,
		Breakdown: models.LatencyBreakdown{
			TCPConnect: latency,
			Total:      latency,
		},
	}, nil
}
