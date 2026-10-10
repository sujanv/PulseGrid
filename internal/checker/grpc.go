package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type GRPCChecker struct{}

func NewGRPCChecker() *GRPCChecker {
	return &GRPCChecker{}
}

func (c *GRPCChecker) Check(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	address := fmt.Sprintf("%s:%d", target.Host, target.Port)
	dialer := &net.Dialer{Timeout: target.Timeout}

	startTime := time.Now()
	// Dial socket with HTTP/2 and TLS negotiation
	tlsConf := &tls.Config{
		ServerName:         target.Host,
		NextProtos:         []string{"h2"},
		InsecureSkipVerify: false,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", address, tlsConf)
	if err != nil {
		// Fallback to plain TCP h2c dial
		rawConn, rawErr := dialer.DialContext(ctx, "tcp", address)
		if rawErr != nil {
			return &models.CheckResult{
				ID:           uuid.New().String(),
				TargetID:     target.ID,
				Timestamp:    time.Now().UTC(),
				Status:       models.StatusDown,
				Latency:      time.Since(startTime),
				ErrorMessage: fmt.Sprintf("gRPC connection failed: %v", rawErr),
			}, nil
		}
		rawConn.Close()
	} else {
		conn.Close()
	}

	latency := time.Since(startTime)
	return &models.CheckResult{
		ID:        uuid.New().String(),
		TargetID:  target.ID,
		Timestamp: time.Now().UTC(),
		Status:    models.StatusUp,
		Latency:   latency,
		Breakdown: models.LatencyBreakdown{
			TCPConnect: latency,
			Total:      latency,
		},
	}, nil
}
