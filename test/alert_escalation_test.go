package test

import (
	"testing"
	"time"

	"github.com/sujanv/PulseGrid/internal/alert"
	"github.com/sujanv/PulseGrid/internal/models"
)

func TestQuietHoursFilter(t *testing.T) {
	filter := alert.NewQuietHoursFilter()
	filter.AddRule(&alert.QuietHoursRule{
		StartHour: 22,
		EndHour:   7,
		TimeZone:  "UTC",
		AllowP0:   true,
	})

	nightTime := time.Date(2026, 10, 11, 23, 30, 0, 0, time.UTC)
	dayTime := time.Date(2026, 10, 11, 14, 0, 0, 0, time.UTC)

	// Non-P0 alert at night -> should suppress
	evNonP0 := &models.AlertEvent{
		EventType: models.AlertFlapping,
		Timestamp: nightTime,
	}
	if !filter.ShouldSuppress(evNonP0) {
		t.Errorf("expected non-P0 event to be suppressed during quiet hours")
	}

	// P0 outage at night -> allowed because AllowP0 is true
	evP0 := &models.AlertEvent{
		EventType: models.AlertOutage,
		Timestamp: nightTime,
	}
	if filter.ShouldSuppress(evP0) {
		t.Errorf("critical outage should not be suppressed")
	}

	// Day event -> never suppressed
	evDay := &models.AlertEvent{
		EventType: models.AlertFlapping,
		Timestamp: dayTime,
	}
	if filter.ShouldSuppress(evDay) {
		t.Errorf("daytime alert should not be suppressed")
	}
}
