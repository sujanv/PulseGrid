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

type TLSChecker struct{}

func NewTLSChecker() *TLSChecker {
	return &TLSChecker{}
}

func (c *TLSChecker) Check(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	host := target.Host
	port := target.Port
	if port <= 0 {
		port = 443
	}
	address := fmt.Sprintf("%s:%d", host, port)

	dialer := &net.Dialer{}
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	startTime := time.Now()
	rawConn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(startTime),
			ErrorMessage: fmt.Sprintf("TCP dial failed to %s: %v", address, err),
		}, nil
	}
	defer rawConn.Close()

	tlsConn := tls.Client(rawConn, tlsConfig)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(startTime),
			ErrorMessage: fmt.Sprintf("TLS handshake failed: %v", err),
		}, nil
	}

	totalLatency := time.Since(startTime)
	certs := tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      totalLatency,
			ErrorMessage: "no peer certificates found",
		}, nil
	}

	leaf := certs[0]
	now := time.Now()
	daysLeft := int(leaf.NotAfter.Sub(now).Hours() / 24)

	status := models.StatusUp
	var errMsg string

	if now.After(leaf.NotAfter) {
		status = models.StatusDown
		errMsg = fmt.Sprintf("certificate expired on %s", leaf.NotAfter.Format(time.RFC3339))
	} else if daysLeft < 14 {
		status = models.StatusDegraded
		errMsg = fmt.Sprintf("certificate expires soon: %d days left (%s)", daysLeft, leaf.NotAfter.Format("2006-01-02"))
	}

	return &models.CheckResult{
		ID:                  uuid.New().String(),
		TargetID:            target.ID,
		Timestamp:           now.UTC(),
		Status:              status,
		Latency:             totalLatency,
		ErrorMessage:        errMsg,
		CertificateDaysLeft: daysLeft,
		Breakdown: models.LatencyBreakdown{
			TLSHandshake: totalLatency,
			Total:        totalLatency,
		},
	}, nil
}
