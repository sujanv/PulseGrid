package test

import (
	"context"
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/engine"
	"github.com/sujanv/PulseGrid/internal/models"
)

func TestIncidentDetectorTransitions(t *testing.T) {
	var alertCount int
	detector := engine.NewIncidentDetector(nil, func(event *models.AlertEvent) {
		alertCount++
	})

	target := &models.Target{
		ID:               "tgt-1",
		Name:             "API Endpoint",
		FailureThreshold: 2,
		SuccessThreshold: 2,
	}

	// 1st failure - should not trigger DOWN yet
	res1 := &models.CheckResult{
		Status:       models.StatusDown,
		ErrorMessage: "500 Internal Error",
		Timestamp:    time.Now(),
	}
	inc1, _ := detector.ProcessResult(context.Background(), target, res1)
	if inc1 != nil {
		t.Errorf("incident should not open after 1 failure when threshold is 2")
	}

	// 2nd failure - threshold reached -> triggers incident
	res2 := &models.CheckResult{
		Status:       models.StatusDown,
		ErrorMessage: "500 Internal Error",
		Timestamp:    time.Now(),
	}
	inc2, alert := detector.ProcessResult(context.Background(), target, res2)
	if inc2 == nil {
		t.Fatalf("incident should have been opened")
	}
	if alert == nil || alert.EventType != models.AlertOutage {
		t.Errorf("expected AlertOutage event")
	}

	// 1st success - should remain DOWN
	res3 := &models.CheckResult{Status: models.StatusUp, Timestamp: time.Now()}
	_, _ = detector.ProcessResult(context.Background(), target, res3)
	if detector.GetTargetStatus(target.ID) != models.StatusDown {
		t.Errorf("target should still be DOWN after 1 recovery check")
	}

	// 2nd success - recovery threshold reached -> resolves
	res4 := &models.CheckResult{Status: models.StatusUp, Timestamp: time.Now()}
	_, alertRec := detector.ProcessResult(context.Background(), target, res4)
	if alertRec == nil || alertRec.EventType != models.AlertRecovery {
		t.Errorf("expected AlertRecovery event")
	}
	if detector.GetTargetStatus(target.ID) != models.StatusUp {
		t.Errorf("target should be UP now")
	}
}
