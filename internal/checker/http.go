package checker

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type HTTPChecker struct {
	client *http.Client
}

func NewHTTPChecker(transportTimeout time.Duration) *HTTPChecker {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   transportTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	return &HTTPChecker{
		client: &http.Client{
			Transport: transport,
			Timeout:   0, // per-request context handles the timeout
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

func (c *HTTPChecker) Check(ctx context.Context, target *models.Target) (*models.CheckResult, error) {
	var (
		dnsStart, dnsEnd         time.Time
		connStart, connEnd       time.Time
		tlsStart, tlsEnd         time.Time
		gotFirstByte             time.Time
		resolvedIPs              []string
		dnsDuration              time.Duration
		connDuration             time.Duration
		tlsDuration              time.Duration
		ttfbDuration             time.Duration
		transferDuration         time.Duration
	)

	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dnsEnd = time.Now()
			dnsDuration = dnsEnd.Sub(dnsStart)
			for _, addr := range info.Addrs {
				resolvedIPs = append(resolvedIPs, addr.String())
			}
		},
		ConnectStart: func(network, addr string) {
			connStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			connEnd = time.Now()
			connDuration = connEnd.Sub(connStart)
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			tlsEnd = time.Now()
			tlsDuration = tlsEnd.Sub(tlsStart)
		},
		GotFirstResponseByte: func() {
			gotFirstByte = time.Now()
		},
	}

	traceCtx := httptrace.WithClientTrace(ctx, trace)

	method := string(target.Method)
	if method == "" {
		method = "GET"
	}

	var reqBody io.Reader
	if target.Body != "" {
		reqBody = strings.NewReader(target.Body)
	}

	req, err := http.NewRequestWithContext(traceCtx, method, target.URL, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "PulseGrid-Uptime-Monitor/1.0 (+https://github.com/sujanv/PulseGrid)")
	for k, v := range target.Headers {
		req.Header.Set(k, v)
	}

	startTime := time.Now()
	resp, err := c.client.Do(req)
	if err != nil {
		totalDuration := time.Since(startTime)
		return &models.CheckResult{
			ID:           uuid.New().String(),
			TargetID:     target.ID,
			Timestamp:    time.Now().UTC(),
			Status:       models.StatusDown,
			Latency:      totalDuration,
			ErrorMessage: err.Error(),
			Breakdown: models.LatencyBreakdown{
				DNSLookup:    dnsDuration,
				TCPConnect:   connDuration,
				TLSHandshake: tlsDuration,
				Total:        totalDuration,
			},
			ResolvedIPs: resolvedIPs,
		}, nil
	}
	defer resp.Body.Close()

	if !gotFirstByte.IsZero() {
		ttfbDuration = gotFirstByte.Sub(startTime)
	}

	transferStart := time.Now()
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // limit to 1MB
	transferDuration = time.Since(transferStart)
	totalDuration := time.Since(startTime)

	status := models.StatusUp
	var errMsg string

	// Validate status code
	expectedStatus := false
	for _, code := range target.ExpectedStatusCodes {
		if resp.StatusCode == code {
			expectedStatus = true
			break
		}
	}
	if !expectedStatus {
		status = models.StatusDown
		errMsg = http.StatusText(resp.StatusCode)
	}

	return &models.CheckResult{
		ID:            uuid.New().String(),
		TargetID:      target.ID,
		Timestamp:     time.Now().UTC(),
		Status:        status,
		StatusCode:    resp.StatusCode,
		ResponseBytes: int64(len(bodyBytes)),
		Latency:       totalDuration,
		ErrorMessage:  errMsg,
		Breakdown: models.LatencyBreakdown{
			DNSLookup:       dnsDuration,
			TCPConnect:      connDuration,
			TLSHandshake:    tlsDuration,
			TTFB:            ttfbDuration,
			ContentTransfer: transferDuration,
			Total:           totalDuration,
		},
		ResolvedIPs: resolvedIPs,
	}, nil
}
