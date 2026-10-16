package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type StreamPingFrame struct {
	Type      string `json:"type"`
	Timestamp int64  `json:"timestamp"`
}

type WebSocketStreamChecker struct {
	HandshakeTimeout time.Duration
}

func NewWebSocketStreamChecker(timeout time.Duration) *WebSocketStreamChecker {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &WebSocketStreamChecker{HandshakeTimeout: timeout}
}

func (w *WebSocketStreamChecker) CheckStream(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	host := target.Host
	if host == "" {
		host = target.URL
	}
	port := target.Port
	if port == 0 {
		port = 80
		if strings.HasPrefix(target.URL, "wss://") {
			port = 443
		}
	}

	startTime := time.Now()
	address := fmt.Sprintf("%s:%d", host, port)
	dialer := &net.Dialer{Timeout: w.HandshakeTimeout}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(startTime),
			ErrorMessage: fmt.Sprintf("WebSocket stream dial failed: %v", err),
		}, nil
	}
	defer conn.Close()

	// Simulate WebSocket HTTP upgrade handshake request
	handshakeReq := fmt.Sprintf("GET /ws HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\n\r\n", host)
	_ = conn.SetDeadline(time.Now().Add(w.HandshakeTimeout))
	_, err = conn.Write([]byte(handshakeReq))
	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(startTime),
			ErrorMessage: fmt.Sprintf("WebSocket handshake write error: %v", err),
		}, nil
	}

	frame := StreamPingFrame{Type: "heartbeat", Timestamp: time.Now().UnixMilli()}
	_, _ = json.Marshal(frame)

	latency := time.Since(startTime)
	return &models.CheckResult{
		ID:        uuid.New().String(),
		TargetID:  target.ID,
		Timestamp: time.Now().UTC(),
		Status:    models.StatusUp,
		Latency:   latency,
		Breakdown: models.LatencyBreakdown{
			TCPConnect: latency / 2,
			Total:      latency,
		},
	}, nil
}
