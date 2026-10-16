package test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/alert"
	"github.com/sujanv/PulseGrid/internal/checker"
	"github.com/sujanv/PulseGrid/internal/exporter"
	"github.com/sujanv/PulseGrid/internal/integrations"
	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/network"
	"github.com/sujanv/PulseGrid/internal/pki"
	"github.com/sujanv/PulseGrid/internal/synthetic"
	"github.com/sujanv/PulseGrid/internal/tracing"
	"github.com/sujanv/PulseGrid/internal/web"
)

func TestGeoDNSResolver(t *testing.T) {
	resolver := network.NewGeoDNSResolver()
	resolver.RegisterPolicy("cdn.example.com", []network.GeoDNSRegionPolicy{
		{Region: "us-east", ExpectedPrefixes: []string{"198.51."}},
		{Region: "eu-west", ExpectedPrefixes: []string{"192.0.2."}},
	})

	sim := map[string][]string{
		"us-east": {"198.51.100.1"},
		"eu-west": {"192.0.2.1"},
	}

	res := resolver.VerifyGeoSteering(context.Background(), "cdn.example.com", sim)
	if !res.Healthy {
		t.Errorf("expected GeoDNS resolution to be healthy and policy compliant")
	}
}

func TestWebSocketStreamChecker(t *testing.T) {
	checker := checker.NewWebSocketStreamChecker(2 * time.Second)
	target := &models.Target{
		ID:      "tgt-ws",
		Host:    "127.0.0.1",
		Port:    8080,
		Timeout: 2 * time.Second,
	}

	// Dialing arbitrary closed port: should return status DOWN gracefully
	res, err := checker.CheckStream(context.Background(), target)
	if err != nil {
		t.Fatalf("unexpected checker error: %v", err)
	}
	if res.Status != models.StatusDown {
		t.Errorf("expected status DOWN for unreachable stream port")
	}
}

func TestACMEManager(t *testing.T) {
	mgr := pki.NewACMEManager()
	cert, err := mgr.RequestCertificate("api.pulsegrid.io", pki.ChallengeHTTP01)
	if err != nil {
		t.Fatalf("failed to request certificate: %v", err)
	}
	if cert.Status != pki.CertStatusIssued {
		t.Errorf("expected issued cert status, got: %s", cert.Status)
	}

	needed, _ := mgr.CheckRenewalNeeded("api.pulsegrid.io", 30)
	if needed {
		t.Errorf("new cert should not need immediate renewal")
	}
}

func TestWebVitalsCalculator(t *testing.T) {
	calc := synthetic.NewWebVitalsCalculator()

	// Good report
	goodRep := calc.Calculate("https://pulsegrid.io", 1800*time.Millisecond, 40*time.Millisecond, 150*time.Millisecond, 0.05)
	if !goodRep.OverallPass || goodRep.LCPRating != synthetic.RatingGood {
		t.Errorf("expected Good Core Web Vitals score")
	}

	// Poor LCP report
	poorRep := calc.Calculate("https://slow.pulsegrid.io", 4500*time.Millisecond, 40*time.Millisecond, 150*time.Millisecond, 0.05)
	if poorRep.OverallPass || poorRep.LCPRating != synthetic.RatingPoor {
		t.Errorf("expected Poor rating for 4.5s LCP")
	}
}

func TestJitterProfilerRFC1889(t *testing.T) {
	profiler := network.NewJitterProfiler()
	now := time.Now()

	samples := []network.PacketProbeSample{
		{Sequence: 1, SendTime: now, RecvTime: now.Add(20 * time.Millisecond), RTT: 20 * time.Millisecond},
		{Sequence: 2, SendTime: now.Add(100 * time.Millisecond), RecvTime: now.Add(125 * time.Millisecond), RTT: 25 * time.Millisecond},
		{Sequence: 3, SendTime: now.Add(200 * time.Millisecond), RecvTime: now.Add(222 * time.Millisecond), RTT: 22 * time.Millisecond},
	}

	res := profiler.ComputeJitter("edge.pulsegrid.io", samples)
	if res.PacketsReceived != 3 || res.PacketLossPct != 0.0 {
		t.Errorf("expected 3 packets received with 0 loss")
	}
	if res.RFC1889Jitter <= 0 {
		t.Errorf("expected calculated non-zero jitter")
	}
}

func TestRemoteWriteFormatter(t *testing.T) {
	client := exporter.NewRemoteWriteClient("")
	res := &models.CheckResult{
		TargetID:  "tgt-1",
		Timestamp: time.Now(),
		Status:    models.StatusUp,
		Latency:   45 * time.Millisecond,
	}

	series := client.BuildTimeSeries(res)
	if len(series) != 2 {
		t.Errorf("expected 2 metrics series (up & latency)")
	}

	debugOut := client.FormatDebugPayload(series)
	if !strings.Contains(debugOut, "pulsegrid_target_up") {
		t.Errorf("debug output missing target_up metric: %s", debugOut)
	}
}

func TestOnCallRotationScheduler(t *testing.T) {
	mgr := alert.NewOnCallRotationManager()
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	sched := &alert.OnCallSchedule{
		ID:        "primary-rotation",
		Name:      "SRE Primary On-Call",
		ShiftType: alert.ShiftWeekly,
		StartDate: start,
		Engineers: []alert.OnCallEngineer{
			{ID: "eng-1", Name: "Alice", Email: "alice@pulsegrid.io"},
			{ID: "eng-2", Name: "Bob", Email: "bob@pulsegrid.io"},
		},
	}
	mgr.RegisterSchedule(sched)

	// Week 1 -> Alice
	active1, err := mgr.GetActiveOnCall("primary-rotation", start.Add(2*24*time.Hour))
	if err != nil || active1.Name != "Alice" {
		t.Errorf("expected Alice for week 1, got %v", active1)
	}

	// Week 2 -> Bob
	active2, err := mgr.GetActiveOnCall("primary-rotation", start.Add(8*24*time.Hour))
	if err != nil || active2.Name != "Bob" {
		t.Errorf("expected Bob for week 2, got %v", active2)
	}
}

func TestBadgeGeneratorSVG(t *testing.T) {
	gen := web.NewBadgeGenerator()
	svg := gen.GenerateSVG(models.StatusUp, 35*time.Millisecond, 99.99)
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "uptime") {
		t.Errorf("invalid SVG badge output: %s", svg)
	}
}

func TestGrafanaDashboardExport(t *testing.T) {
	gen := integrations.NewGrafanaDashboardGenerator()
	jsonStr, err := gen.ExportDashboardJSON()
	if err != nil || !strings.Contains(jsonStr, "PulseGrid Enterprise Monitoring") {
		t.Errorf("invalid Grafana dashboard export: %v", err)
	}
}

func TestSpanLinker(t *testing.T) {
	linker := tracing.NewSpanLinker()
	linker.LinkOutageSpan("tgt-payment", "trace-100", "span-500", "UPSTREAM_TIMEOUT")

	chain, ok := linker.GetChain("tgt-payment")
	if !ok || len(chain.Links) != 1 {
		t.Errorf("expected 1 linked span in outage chain")
	}
}
