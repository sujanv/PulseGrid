package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/assertion"
	"github.com/sujanv/PulseGrid/internal/models"
)

type HTTPChecker struct {
	client          *http.Client
	assertionEngine *assertion.Engine
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
			Timeout:   0,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
		assertionEngine: assertion.NewEngine(),
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

	req.Header.Set("User-Agent", "PulseGrid-Uptime-Monitor/2.0 (+https://github.com/sujanv/PulseGrid)")
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
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024)) // limit to 2MB
	transferDuration = time.Since(transferStart)
	totalDuration := time.Since(startTime)

	status := models.StatusUp
	var errMsg string

	// 1. Status Code Validation
	expectedStatus := false
	if len(target.ExpectedStatusCodes) == 0 {
		expectedStatus = (resp.StatusCode >= 200 && resp.StatusCode < 300)
	} else {
		for _, code := range target.ExpectedStatusCodes {
			if resp.StatusCode == code {
				expectedStatus = true
				break
			}
		}
	}

	if !expectedStatus {
		status = models.StatusDown
		errMsg = fmt.Sprintf("status code %d not in expected list", resp.StatusCode)
	} else if target.ExpectedKeyword != "" {
		// 2. Keyword Assertion
		bodyStr := string(bodyBytes)
		if !strings.Contains(bodyStr, target.ExpectedKeyword) {
			status = models.StatusDown
			errMsg = fmt.Sprintf("expected keyword '%s' not found in response body", target.ExpectedKeyword)
		}
	}

	// 3. Advanced Assertion Rules (JSONPath, Regex, Header)
	if status == models.StatusUp && len(target.Assertions) > 0 {
		var convertedRules []assertion.AssertionRule
		for _, a := range target.Assertions {
			convertedRules = append(convertedRules, assertion.AssertionRule{
				Type:     assertion.RuleType(a.Type),
				Property: a.Property,
				Operator: assertion.Operator(a.Operator),
				Expected: a.Expected,
			})
		}
		if passed, assertionErr := c.assertionEngine.EvaluateAll(bodyBytes, resp.Header, convertedRules); !passed {
			status = models.StatusDown
			errMsg = assertionErr
		}
	}

	// 4. Latency SLA Evaluation -> Degraded Status
	if status == models.StatusUp && target.DegradedThresholdMs > 0 {
		if totalDuration > time.Duration(target.DegradedThresholdMs)*time.Millisecond {
			status = models.StatusDegraded
			errMsg = fmt.Sprintf("response latency (%v) breached SLA threshold of %dms", totalDuration, target.DegradedThresholdMs)
		}
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
