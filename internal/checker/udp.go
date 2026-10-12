package checker

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type UDPChecker struct{}

func NewUDPChecker() *UDPChecker {
	return &UDPChecker{}
}

func (u *UDPChecker) Check(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	address := fmt.Sprintf("%s:%d", target.Host, target.Port)
	dialer := &net.Dialer{Timeout: target.Timeout}

	start := time.Now()
	conn, err := dialer.DialContext(ctx, "udp", address)
	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(start),
			ErrorMessage: fmt.Sprintf("UDP dial failed: %v", err),
		}, nil
	}
	defer conn.Close()

	// Send probe payload
	probe := []byte("pulsegrid-probe\n")
	_ = conn.SetDeadline(time.Now().Add(target.Timeout))
	_, err = conn.Write(probe)
	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(start),
			ErrorMessage: fmt.Sprintf("UDP write error: %v", err),
		}, nil
	}

	latency := time.Since(start)
	return &models.CheckResult{
		ID:        uuid.New().String(),
		TargetID:  target.ID,
		Timestamp: time.Now().UTC(),
		Status:    models.StatusUp,
		Latency:   latency,
		Breakdown: models.LatencyBreakdown{
			Total: latency,
		},
	}, nil
}
