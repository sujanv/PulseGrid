package test

import (
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/remediation"
)

func TestRemediationCircuitCooldown(t *testing.T) {
	circuit := remediation.NewCircuitBreaker()
	targetID := "tgt-restart-pod"
	cooldown := 1 * time.Second

	if !circuit.Allow(targetID, cooldown) {
		t.Errorf("1st trigger should be allowed")
	}

	if circuit.Allow(targetID, cooldown) {
		t.Errorf("2nd trigger should be blocked by cooldown")
	}

	time.Sleep(1100 * time.Millisecond)

	if !circuit.Allow(targetID, cooldown) {
		t.Errorf("trigger after cooldown should be allowed")
	}
}
