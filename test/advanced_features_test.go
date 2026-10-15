package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/alert"
	"github.com/sujanv/PulseGrid/internal/chaos"
	"github.com/sujanv/PulseGrid/internal/consensus"
	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/rca"
	"github.com/sujanv/PulseGrid/internal/synthetic"
)

func TestDOMMatcher(t *testing.T) {
	matcher := synthetic.NewDOMMatcher()
	html := `<html><body><div id="status" class="alert success">Service Operational</div><title>PulseGrid Status</title></body></html>`

	if !matcher.ContainsText(html, "Service Operational") {
		t.Errorf("expected HTML to contain 'Service Operational'")
	}
	if !matcher.HasElementWithID(html, "status") {
		t.Errorf("expected HTML to have element with ID 'status'")
	}
	if !matcher.HasElementWithClass(html, "success") {
		t.Errorf("expected HTML to have element with class 'success'")
	}

	titles := matcher.ExtractTagContent(html, "title")
	if len(titles) == 0 || titles[0] != "PulseGrid Status" {
		t.Errorf("expected title 'PulseGrid Status', got %v", titles)
	}
}

func TestGeoMatrixConsensus(t *testing.T) {
	evaluator := consensus.NewGeoMatrixEvaluator(0.60)

	// Case 1: Healthy
	probesAllUp := []consensus.RegionProbeStatus{
		{Region: "us-east", IsUp: true, Latency: 25 * time.Millisecond},
		{Region: "eu-west", IsUp: true, Latency: 35 * time.Millisecond},
		{Region: "ap-south", IsUp: true, Latency: 45 * time.Millisecond},
	}
	evalUp := evaluator.Evaluate("tgt-1", probesAllUp)
	if evalUp.Scope != consensus.ScopeHealthy {
		t.Errorf("expected ScopeHealthy, got %v", evalUp.Scope)
	}

	// Case 2: Localized edge failure (1 out of 3 failing = 33% < 60%)
	probesEdgeDown := []consensus.RegionProbeStatus{
		{Region: "us-east", IsUp: true, Latency: 25 * time.Millisecond},
		{Region: "eu-west", IsUp: false, Latency: 500 * time.Millisecond},
		{Region: "ap-south", IsUp: true, Latency: 45 * time.Millisecond},
	}
	evalEdge := evaluator.Evaluate("tgt-1", probesEdgeDown)
	if evalEdge.Scope != consensus.ScopeLocalized {
		t.Errorf("expected ScopeLocalized, got %v", evalEdge.Scope)
	}

	// Case 3: Global outage (2 out of 3 failing = 66% >= 60%)
	probesGlobalDown := []consensus.RegionProbeStatus{
		{Region: "us-east", IsUp: false, Latency: 500 * time.Millisecond},
		{Region: "eu-west", IsUp: false, Latency: 500 * time.Millisecond},
		{Region: "ap-south", IsUp: true, Latency: 45 * time.Millisecond},
	}
	evalGlobal := evaluator.Evaluate("tgt-1", probesGlobalDown)
	if evalGlobal.Scope != consensus.ScopeGlobal {
		t.Errorf("expected ScopeGlobal, got %v", evalGlobal.Scope)
	}
}

func TestFlapDampenerHysteresis(t *testing.T) {
	dampener := consensus.NewFlapDampener(50.0, 20.0, 10*time.Second)
	now := time.Now()

	// Initial stable
	flapping, suppress := dampener.ProcessStatus("tgt-flapping", true, now)
	if flapping || suppress {
		t.Errorf("initially should not be flapping")
	}

	// Rapid oscillations
	dampener.ProcessStatus("tgt-flapping", false, now.Add(1*time.Second))
	flapping, suppress = dampener.ProcessStatus("tgt-flapping", true, now.Add(2*time.Second))
	if !flapping || !suppress {
		t.Errorf("after rapid oscillations target should be flapping")
	}
}

func TestFaultInjectorAndBlastRadius(t *testing.T) {
	injector := chaos.NewFaultInjector()
	injector.RegisterRule(&chaos.FaultRule{
		ID:          "rule-1",
		TargetID:    "staging-tgt",
		Type:        chaos.FaultPacketLoss,
		Probability: 1.0,
		Active:      true,
	})

	target := &models.Target{
		ID:      "staging-tgt",
		Timeout: 5 * time.Second,
	}
	orig := &models.CheckResult{
		Status:  models.StatusUp,
		Latency: 10 * time.Millisecond,
	}

	mutated, applied := injector.ApplyFault(context.Background(), target, orig)
	if !applied || mutated.Status != models.StatusDown {
		t.Errorf("expected chaos fault to turn status DOWN")
	}

	scheduler := chaos.NewChaosScheduler(injector)
	// Try launching in production: must fail
	errProd := scheduler.LaunchExperiment(&chaos.ChaosExperiment{
		ID:        "exp-prod",
		TargetEnv: "production",
	})
	if errProd == nil {
		t.Errorf("expected error when launching chaos experiment in production")
	}
}

func TestWebhookHMACSigningAndVerification(t *testing.T) {
	secret := "super-secure-webhook-secret-key-1234"
	signer := alert.NewWebhookSigner(secret)
	payload := []byte(`{"event":"outage","target":"payment-api"}`)
	now := time.Now()

	sig, ts := signer.SignPayload(payload, now)
	if sig == "" || ts == "" {
		t.Fatalf("expected valid signature and timestamp")
	}

	// Verify valid signature
	if !signer.VerifySignature(payload, sig, ts, 5*time.Minute) {
		t.Errorf("expected signature to be verified successfully")
	}

	// Verify tampered payload fails
	tampered := []byte(`{"event":"outage","target":"payment-api-TAMPERED"}`)
	if signer.VerifySignature(tampered, sig, ts, 5*time.Minute) {
		t.Errorf("expected verification to fail for tampered payload")
	}
}

func TestPostMortemAndReliabilityMetrics(t *testing.T) {
	gen := rca.NewPostMortemGenerator()
	now := time.Now().UTC()

	incident := &rca.IncidentData{
		IncidentID:     "INC-999",
		TargetName:     "Auth Service",
		TargetURL:      "https://auth.example.com",
		RootCause:      rca.CauseCertExpired,
		Recommendation: "Automate ACME certificate renewal with cert-manager.",
		StartedAt:      now.Add(-10 * time.Minute),
		ResolvedAt:     now,
		Timeline: []rca.IncidentTimelineEntry{
			{Timestamp: now.Add(-10 * time.Minute), Description: "TLS handshake failed: cert expired", Severity: "CRITICAL", Source: "PulseGrid"},
		},
	}

	md := gen.GenerateMarkdown(incident)
	if !strings.Contains(md, "INC-999") || !strings.Contains(md, "SSL_CERTIFICATE_EXPIRED") {
		t.Errorf("postmortem markdown missing key details: %s", md)
	}

	calc := rca.NewReliabilityMetricsCalculator()
	metrics := calc.Calculate("INC-999", now.Add(-10*time.Minute), now.Add(-9*time.Minute), now, 99.9)
	if metrics.TotalDowntimeMinutes < 9.9 || metrics.TotalDowntimeMinutes > 10.1 {
		t.Errorf("expected ~10 min downtime, got %f", metrics.TotalDowntimeMinutes)
	}
}

func TestScenarioRunnerExecution(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/step1":
			http.SetCookie(w, &http.Cookie{Name: "session_id", Value: "token-abc-123"})
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Step 1 Success"))
		case "/step2":
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value != "token-abc-123" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Step 2 Authenticated"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	runner := synthetic.NewScenarioRunner()
	steps := []synthetic.ScenarioStep{
		{
			Name:           "Login Step",
			Method:         http.MethodGet,
			URL:            ts.URL + "/step1",
			ExpectedStatus: http.StatusOK,
			Timeout:        3 * time.Second,
		},
		{
			Name:           "Authenticated Profile Step",
			Method:         http.MethodGet,
			URL:            ts.URL + "/step2",
			ExpectedStatus: http.StatusOK,
			Timeout:        3 * time.Second,
		},
	}

	res, err := runner.Execute(context.Background(), "scen-1", "Test Scenario", steps)
	if err != nil {
		t.Fatalf("scenario execution error: %v", err)
	}
	if !res.Success {
		t.Errorf("expected scenario to succeed with persisted cookies across steps")
	}
}
