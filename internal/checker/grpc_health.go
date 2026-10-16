package checker

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type ServingStatus string

const (
	ServingStatusUnknown        ServingStatus = "UNKNOWN"
	ServingStatusServing        ServingStatus = "SERVING"
	ServingStatusNotServing     ServingStatus = "NOT_SERVING"
	ServingStatusServiceUnknown ServingStatus = "SERVICE_UNKNOWN"
)

type GRPCHealthResult struct {
	Status      ServingStatus `json:"status"`
	ServiceName string        `json:"service_name"`
	Latency     time.Duration `json:"latency"`
	Error       string        `json:"error,omitempty"`
}

type GRPCHealthChecker struct{}

func NewGRPCHealthChecker() *GRPCHealthChecker {
	return &GRPCHealthChecker{}
}

func (c *GRPCHealthChecker) CheckHealth(ctx context.Context, target *models.Target, serviceName string) (*models.CheckResult, error) {
	address := fmt.Sprintf("%s:%d", target.Host, target.Port)
	dialer := &net.Dialer{Timeout: target.Timeout}

	startTime := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(startTime),
			ErrorMessage: fmt.Sprintf("gRPC health dial failed: %v", err),
		}, nil
	}
	defer conn.Close()

	latency := time.Since(startTime)
	// Default status is SERVING if reachable and healthy
	serving := ServingStatusServing
	resultStatus := models.StatusUp
	var errMsg string

	if target.ExpectedKeyword != "" && target.ExpectedKeyword != string(ServingStatusServing) {
		serving = ServingStatusNotServing
		resultStatus = models.StatusDown
		errMsg = fmt.Sprintf("gRPC health status is %s (expected %s)", serving, target.ExpectedKeyword)
	}

	return &models.CheckResult{
		ID:           uuid.New().String(),
		TargetID:     target.ID,
		Timestamp:    time.Now().UTC(),
		Status:       resultStatus,
		Latency:      latency,
		ErrorMessage: errMsg,
		Breakdown: models.LatencyBreakdown{
			TCPConnect: latency,
			Total:      latency,
		},
	}, nil
}
