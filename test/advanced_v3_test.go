package test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/alert"
	"github.com/sujanv/PulseGrid/internal/api"
	"github.com/sujanv/PulseGrid/internal/canary"
	"github.com/sujanv/PulseGrid/internal/checker"
	"github.com/sujanv/PulseGrid/internal/integrations"
	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/network"
	"github.com/sujanv/PulseGrid/internal/sampling"
	"github.com/sujanv/PulseGrid/internal/synthetic"
)

func TestCertInspectorSANAndExpiry(t *testing.T) {
	inspector := checker.NewCertInspector()

	// Connect to test TLS server to get a real certificate chain
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	conn, err := tls.Dial("tcp", server.Listener.Addr().String(), &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		t.Fatalf("failed tls dial: %v", err)
	}
	defer conn.Close()

	chain := conn.ConnectionState().PeerCertificates
	rep, err := inspector.Inspect(chain, "127.0.0.1")
	if err != nil {
		t.Fatalf("inspection error: %v", err)
	}

	if rep.ChainLength == 0 {
		t.Errorf("expected chain length > 0")
	}
}

func TestHARExporter(t *testing.T) {
	exporter := synthetic.NewHARExporter()
	scenResult := &synthetic.ScenarioResult{
		ScenarioID:    "test-scen",
		ScenarioName:  "Synthetic Test",
		Success:       true,
		ExecutedAt:    time.Now().UTC(),
		TotalDuration: 100 * time.Millisecond,
		StepResults: []synthetic.ScenarioStepResult{
			{
				StepName:   "Get User",
				URL:        "https://api.example.com/user",
				StatusCode: 200,
				Success:    true,
				Waterfall: synthetic.StepPerformanceWaterfall{
					DNSDuration:      5 * time.Millisecond,
					ConnectDuration:  10 * time.Millisecond,
					TLSDuration:      15 * time.Millisecond,
					TTFB:             50 * time.Millisecond,
					TransferDuration: 20 * time.Millisecond,
					TotalDuration:    100 * time.Millisecond,
				},
			},
		},
	}

	har := exporter.ExportScenarioToHAR(scenResult)
	if har.Log.Version != "1.2" {
		t.Errorf("expected HAR 1.2, got %s", har.Log.Version)
	}
	if len(har.Log.Entries) != 1 {
		t.Errorf("expected 1 HAR entry, got %d", len(har.Log.Entries))
	}

	jsonStr, err := exporter.ToJSON(har)
	if err != nil || len(jsonStr) == 0 {
		t.Errorf("expected valid HAR JSON output")
	}
}

func TestAdaptiveSampler(t *testing.T) {
	sampler := sampling.NewAdaptiveSampler(0.20, 200*time.Millisecond)

	// Failing target -> 100% sampling
	decDown := sampler.Evaluate("tgt-1", models.StatusDown, 50*time.Millisecond)
	if !decDown.ShouldSample || decDown.SampleRate != 1.0 {
		t.Errorf("failing target must have 100%% sample rate")
	}

	// Degraded target (>200ms) -> 100% sampling
	decDeg := sampler.Evaluate("tgt-2", models.StatusUp, 350*time.Millisecond)
	if !decDeg.ShouldSample || decDeg.SampleRate != 1.0 {
		t.Errorf("degraded target must have 100%% sample rate")
	}
}

func TestAlertDeduplicator(t *testing.T) {
	dedup := alert.NewAlertDeduplicator(1*time.Minute, 3)
	domain := alert.FailureDomain{
		Provider: "AWS",
		Region:   "us-east-1",
		Zone:     "us-east-1a",
		Service:  "api-gateway",
	}

	evt1 := &models.AlertEvent{TargetID: "tgt-1"}
	isNew, grp := dedup.Deduplicate(evt1, domain)
	if !isNew || grp.Count != 1 {
		t.Errorf("first alert should be new")
	}

	evt2 := &models.AlertEvent{TargetID: "tgt-2"}
	isNew, grp = dedup.Deduplicate(evt2, domain)
	if isNew {
		t.Errorf("second alert within window should be deduplicated")
	}

	evt3 := &models.AlertEvent{TargetID: "tgt-3"}
	isNew, grp = dedup.Deduplicate(evt3, domain)
	if !isNew || !grp.IsCluster {
		t.Errorf("third alert should trigger cluster escalation")
	}
}

func TestCanaryEvaluator(t *testing.T) {
	evaluator := canary.NewCanaryEvaluator(0.02, 50*time.Millisecond)

	base := canary.TargetMetricsSummary{
		ErrorRate:  0.005,
		P95Latency: 50 * time.Millisecond,
	}

	// Canary with high errors -> ROLLBACK
	badCanary := canary.TargetMetricsSummary{
		ErrorRate:  0.05, // 5% error rate vs 0.5%
		P95Latency: 55 * time.Millisecond,
	}
	resBad := evaluator.Evaluate("base", "canary", base, badCanary)
	if resBad.Decision != canary.DecisionRollback {
		t.Errorf("expected ROLLBACK, got %v", resBad.Decision)
	}

	// Good canary -> PROMOTE
	goodCanary := canary.TargetMetricsSummary{
		ErrorRate:  0.004,
		P95Latency: 45 * time.Millisecond,
	}
	resGood := evaluator.Evaluate("base", "canary", base, goodCanary)
	if resGood.Decision != canary.DecisionPromote {
		t.Errorf("expected PROMOTE, got %v", resGood.Decision)
	}
}

func TestChatWebhooks(t *testing.T) {
	evt := &models.AlertEvent{
		TargetID:   "tgt-api",
		TargetName: "Payment Gateway",
		TargetURL:  "https://api.example.com",
		EventType:  "OUTAGE",
		Message:    "503 Service Unavailable",
		Timestamp:  time.Now(),
	}

	teams := integrations.NewTeamsCardFormatter()
	teamsBytes, err := teams.FormatCard(evt)
	if err != nil || len(teamsBytes) == 0 {
		t.Errorf("expected valid Teams Card JSON")
	}

	discord := integrations.NewDiscordEmbedFormatter()
	discordBytes, err := discord.FormatEmbed(evt)
	if err != nil || len(discordBytes) == 0 {
		t.Errorf("expected valid Discord Embed JSON")
	}
}

func TestTokenBucketRateLimiter(t *testing.T) {
	limiter := api.NewTokenBucketLimiter(3.0, 1.0)

	// First 3 requests should pass
	for i := 0; i < 3; i++ {
		allowed, _, _ := limiter.Allow("client-1")
		if !allowed {
			t.Errorf("request %d should be allowed within burst capacity", i)
		}
	}

	// 4th request must be rejected
	allowed, _, retryAfter := limiter.Allow("client-1")
	if allowed {
		t.Errorf("4th request should exceed capacity")
	}
	if retryAfter <= 0 {
		t.Errorf("expected positive retry after duration")
	}
}

func TestAnycastMatrixEvaluator(t *testing.T) {
	matrix := network.NewAnycastMatrixEvaluator()

	initial := []network.AnycastRegionMetric{
		{Region: "us-east-1", ResolvedIP: "198.51.100.1", MeanRTT: 20 * time.Millisecond},
		{Region: "eu-west-1", ResolvedIP: "198.51.100.2", MeanRTT: 30 * time.Millisecond},
	}
	shifts1 := matrix.RecordAndCheckShifts("edge.pulsegrid.io", initial)
	if len(shifts1) > 0 {
		t.Errorf("initial record should not produce shifts")
	}

	// Route shift on us-east-1
	updated := []network.AnycastRegionMetric{
		{Region: "us-east-1", ResolvedIP: "198.51.100.99", MeanRTT: 45 * time.Millisecond},
		{Region: "eu-west-1", ResolvedIP: "198.51.100.2", MeanRTT: 30 * time.Millisecond},
	}
	shifts2 := matrix.RecordAndCheckShifts("edge.pulsegrid.io", updated)
	if len(shifts2) != 1 || shifts2[0].NewIP != "198.51.100.99" {
		t.Errorf("expected 1 detected shift on us-east-1, got %v", shifts2)
	}
}
