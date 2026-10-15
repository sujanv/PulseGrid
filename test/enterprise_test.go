package test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/auth"
	"github.com/sujanv/PulseGrid/internal/network"
	"github.com/sujanv/PulseGrid/internal/rca"
	"github.com/sujanv/PulseGrid/internal/slo"
	"github.com/sujanv/PulseGrid/internal/statuspage"
)

func TestTracerouteAndRouteAnalyzer(t *testing.T) {
	engine := network.NewTracerouteEngine(10, 2*time.Second)
	ctx := context.Background()

	res, err := engine.Trace(ctx, "example.com")
	if err != nil {
		t.Fatalf("unexpected traceroute error: %v", err)
	}
	if len(res.Hops) == 0 {
		t.Errorf("expected hops in traceroute result")
	}

	analyzer := network.NewRouteAnalyzer(50 * time.Millisecond)
	anomalies := analyzer.DetectAnomalies(nil, res)
	_ = anomalies // may or may not have anomalies depending on random latency
}

func TestGoogleSREBurnRateEvaluator(t *testing.T) {
	evaluator := slo.NewBurnRateEvaluator(nil)

	// Normal healthy traffic: 10,000 requests, 1 failure -> burn rate close to 0
	normal := evaluator.Evaluate("tgt-1", 99.9, 1*time.Hour, 10000, 1)
	if normal.IsBreaching {
		t.Errorf("expected normal traffic not to breach burn rate")
	}

	// Critical 14.4x burn rate breach: 10,000 requests, 200 failures
	breach := evaluator.Evaluate("tgt-1", 99.9, 1*time.Hour, 10000, 200)
	if !breach.IsBreaching || breach.AlertSeverity != slo.SeverityCritical {
		t.Errorf("expected critical burn rate breach, got: %+v", breach)
	}
}

func TestErrorBudgetCalculator(t *testing.T) {
	calc := slo.NewErrorBudgetCalculator()
	budget := calc.Calculate("tgt-1", 99.9, 30, 10*time.Minute, 1.0)

	if budget.TotalMinutes != 43200 {
		t.Errorf("expected 43200 total minutes in 30 days, got %f", budget.TotalMinutes)
	}
	if budget.AllowedDowntimeMinutes <= 43.1 || budget.AllowedDowntimeMinutes >= 43.3 {
		t.Errorf("expected ~43.2 allowed minutes for 99.9%%, got %f", budget.AllowedDowntimeMinutes)
	}
	if budget.RemainingDowntimeMinutes <= 33.1 || budget.RemainingDowntimeMinutes >= 33.3 {
		t.Errorf("expected ~33.2 remaining minutes, got %f", budget.RemainingDowntimeMinutes)
	}
}

func TestRBACAndJWTTokenService(t *testing.T) {
	rbac := auth.NewRBACChecker()

	if !rbac.HasPermission(auth.RoleAdmin, auth.PermAlertsManage) {
		t.Errorf("expected admin to have alert management permission")
	}
	if rbac.HasPermission(auth.RoleViewer, auth.PermTargetsWrite) {
		t.Errorf("viewer should not have target write permission")
	}

	tokenSvc := auth.NewTokenService("test-secret-key-32-bytes-long!")
	token, err := tokenSvc.GenerateToken("tenant-42", "user-1", auth.RoleOperator, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate JWT token: %v", err)
	}

	claims, err := tokenSvc.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate JWT token: %v", err)
	}
	if claims.TenantID != "tenant-42" || claims.Role != auth.RoleOperator {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestRSSAndAtomFeedGeneration(t *testing.T) {
	gen := statuspage.NewFeedGenerator("Test Status", "https://status.example.com")
	items := []statuspage.IncidentFeedItem{
		{
			ID:          "inc-101",
			Title:       "Database Connection Latency Spike",
			Description: "Failover in progress.",
			PubDate:     time.Now().UTC(),
			Link:        "https://status.example.com/incidents/101",
		},
	}

	rss := gen.GenerateRSS(items)
	if !strings.Contains(rss, "<rss version="2.0">") || !strings.Contains(rss, "inc-101") {
		t.Errorf("invalid RSS output: %s", rss)
	}

	atom := gen.GenerateAtom(items)
	if !strings.Contains(atom, "http://www.w3.org/2005/Atom") || !strings.Contains(atom, "inc-101") {
		t.Errorf("invalid Atom output: %s", atom)
	}
}

func TestRCADiagnosis(t *testing.T) {
	rcaEngine := rca.NewRCAEngine()

	diagDNS := rcaEngine.Diagnose("tgt-1", "dial tcp: lookup non-existent-api.domain: no such host", 0)
	if diagDNS.RootCause != rca.CauseDNSFailure {
		t.Errorf("expected DNS root cause, got: %v", diagDNS.RootCause)
	}

	diag502 := rcaEngine.Diagnose("tgt-2", "Bad Gateway", 502)
	if diag502.RootCause != rca.CauseGateway502 {
		t.Errorf("expected 502 Bad Gateway root cause, got: %v", diag502.RootCause)
	}
}
