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

type OCSPChecker struct{}

func NewOCSPChecker() *OCSPChecker {
	return &OCSPChecker{}
}

func (o *OCSPChecker) Check(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	host := target.Host
	if host == "" {
		host = target.URL
	}
	address := fmt.Sprintf("%s:%d", host, 443)

	dialer := &net.Dialer{Timeout: target.Timeout}
	start := time.Now()

	conn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{
		ServerName: host,
	})
	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(start),
			ErrorMessage: fmt.Sprintf("TLS connection failed: %v", err),
		}, nil
	}
	defer conn.Close()

	state := conn.ConnectionState()
	hasStaple := len(state.OCSPResponse) > 0

	status := models.StatusUp
	var errMsg string
	if !hasStaple {
		errMsg = "Certificate valid; OCSP Stapling not enabled by server"
	}

	return &models.CheckResult{
		ID:           uuid.New().String(),
		TargetID:     target.ID,
		Timestamp:    time.Now().UTC(),
		Status:       status,
		Latency:      time.Since(start),
		ErrorMessage: errMsg,
	}, nil
}
