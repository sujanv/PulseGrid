package checker

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type PingChecker struct{}

func NewPingChecker() *PingChecker {
	return &PingChecker{}
}

func (c *PingChecker) Check(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	host := target.Host
	if host == "" {
		host = target.URL
	}

	// Standard unprivileged echo via TCP port 7 fallback or fast TCP roundtrip to standard ports
	port := target.Port
	if port <= 0 {
		port = 80 // fallback probe port for host reachability
	}

	address := fmt.Sprintf("%s:%d", host, port)
	dialer := &net.Dialer{
		Timeout: target.Timeout,
	}

	// Run 3 probe iterations to calculate min, avg, max RTT and packet loss
	attempts := 3
	var totalRTT time.Duration
	successCount := 0
	var resolvedIPs []string

	for i := 0; i < attempts; i++ {
		start := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", address)
		elapsed := time.Since(start)

		if err == nil {
			successCount++
			totalRTT += elapsed
			if len(resolvedIPs) == 0 && conn.RemoteAddr() != nil {
				resolvedIPs = append(resolvedIPs, conn.RemoteAddr().String())
			}
			conn.Close()
		}
		time.Sleep(20 * time.Millisecond)
	}

	if successCount == 0 {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      target.Timeout,
			ErrorMessage: fmt.Sprintf("100%% packet loss probing %s", host),
		}, nil
	}

	avgRTT := totalRTT / time.Duration(successCount)
	status := models.StatusUp
	var errMsg string

	if successCount < attempts {
		status = models.StatusDegraded
		lossPct := (float64(attempts-successCount) / float64(attempts)) * 100
		errMsg = fmt.Sprintf("%.0f%% packet loss detected", lossPct)
	}

	return &models.CheckResult{
		ID:          uuid.New().String(),
		TargetID:    target.ID,
		Timestamp:   time.Now().UTC(),
		Status:      status,
		Latency:     avgRTT,
		ErrorMessage: errMsg,
		ResolvedIPs: resolvedIPs,
		Breakdown: models.LatencyBreakdown{
			TCPConnect: avgRTT,
			Total:      avgRTT,
		},
	}, nil
}
