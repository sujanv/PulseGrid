package checker

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type WebSocketChecker struct{}

func NewWebSocketChecker() *WebSocketChecker {
	return &WebSocketChecker{}
}

func (c *WebSocketChecker) Check(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	u, err := url.Parse(target.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid websocket url: %w", err)
	}

	isSecure := (u.Scheme == "wss" || u.Scheme == "https")
	port := u.Port()
	if port == "" {
		if isSecure {
			port = "443"
		} else {
			port = "80"
		}
	}
	hostPort := net.JoinHostPort(u.Hostname(), port)

	dialer := &net.Dialer{Timeout: target.Timeout}
	startTime := time.Now()

	var conn net.Conn
	if isSecure {
		conn, err = tls.DialWithDialer(dialer, "tcp", hostPort, &tls.Config{
			ServerName: u.Hostname(),
		})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", hostPort)
	}

	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(startTime),
			ErrorMessage: fmt.Sprintf("WebSocket dial failed: %v", err),
		}, nil
	}
	defer conn.Close()

	// Generate nonce key
	nonce := make([]byte, 16)
	_, _ = rand.Read(nonce)
	secKey := base64.StdEncoding.EncodeToString(nonce)

	reqPath := u.RequestURI()
	if reqPath == "" {
		reqPath = "/"
	}

	// Send RFC 6455 Handshake request
	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Key: %s\r\nSec-WebSocket-Version: 13\r\n\r\n",
		reqPath, u.Host, secKey)

	_ = conn.SetDeadline(time.Now().Add(target.Timeout))
	if _, err := conn.Write([]byte(req)); err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(startTime),
			ErrorMessage: fmt.Sprintf("WebSocket handshake write failed: %v", err),
		}, nil
	}

	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      time.Since(startTime),
			ErrorMessage: fmt.Sprintf("WebSocket response error: %v", err),
		}, nil
	}

	latency := time.Since(startTime)
	if !strings.Contains(statusLine, "101") && !strings.Contains(statusLine, "200") {
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      latency,
			ErrorMessage: fmt.Sprintf("Invalid handshake response: %s", strings.TrimSpace(statusLine)),
		}, nil
	}

	return &models.CheckResult{
		ID:        uuid.New().String(),
		TargetID:  target.ID,
		Timestamp: time.Now().UTC(),
		Status:    models.StatusUp,
		Latency:   latency,
		Breakdown: models.LatencyBreakdown{
			TTFB:  latency,
			Total: latency,
		},
	}, nil
}
