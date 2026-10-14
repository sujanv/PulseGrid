package test

import (
	"testing"

	"github.com/sujanv/PulseGrid/internal/k8s"
)

func TestIngressAnnotationParser(t *testing.T) {
	parser := k8s.NewIngressAnnotationParser()

	// 1. Not enabled
	_, enabled := parser.ParseAnnotations("my-ingress", "api.example.com", map[string]string{
		"pulsegrid.io/monitor": "false",
	})
	if enabled {
		t.Errorf("expected ingress to not be monitored")
	}

	// 2. Enabled with custom path and interval
	target, enabled := parser.ParseAnnotations("payment-api", "pay.example.com", map[string]string{
		"pulsegrid.io/monitor":  "true",
		"pulsegrid.io/path":     "/health",
		"pulsegrid.io/interval": "15",
	})
	if !enabled || target == nil {
		t.Fatalf("expected target to be created")
	}

	if target.URL != "https://pay.example.com/health" {
		t.Errorf("expected URL https://pay.example.com/health, got %s", target.URL)
	}
	if target.Interval.Seconds() != 15 {
		t.Errorf("expected 15s interval, got %v", target.Interval)
	}
}

func TestCRDControllerReconcile(t *testing.T) {
	controller := k8s.NewController(nil)
	pt := &k8s.ProbeTarget{
		Spec: k8s.ProbeTargetSpec{
			URL:             "https://example.com",
			IntervalSeconds: 30,
			Active:          true,
		},
	}
	pt.Metadata.Name = "test-probe"
	pt.Metadata.Namespace = "default"

	target := controller.Reconcile(pt)
	if target == nil || target.Name != "k8s-test-probe" {
		t.Errorf("reconciliation failed")
	}
}
