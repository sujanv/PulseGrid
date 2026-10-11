package test

import (
	"testing"

	"github.com/sujanv/PulseGrid/internal/anomaly"
)

func TestWelfordAlgorithm(t *testing.T) {
	w := anomaly.NewWelfordStats()
	values := []float64{10.0, 12.0, 11.0, 13.0, 10.0}

	for _, v := range values {
		w.Update(v)
	}

	if w.Count != 5 {
		t.Errorf("expected count 5, got %d", w.Count)
	}

	expectedMean := 11.2
	if (w.Mean - expectedMean) > 0.01 || (expectedMean - w.Mean) > 0.01 {
		t.Errorf("expected mean %.2f, got %.2f", expectedMean, w.Mean)
	}
}

func TestAnomalyDetectorSpike(t *testing.T) {
	detector := anomaly.NewAnomalyDetector(3.0)
	targetID := "tgt-spike"

	// Seed 25 normal measurements around 50ms
	for i := 0; i < 25; i++ {
		detector.Observe(targetID, 50.0)
	}

	// Normal check
	_, isAnomaly := detector.Observe(targetID, 52.0)
	if isAnomaly {
		t.Errorf("52ms should not be flagged as anomaly")
	}

	// 5000ms massive spike
	ev, isSpike := detector.Observe(targetID, 5000.0)
	if !isSpike || ev == nil {
		t.Errorf("expected 5000ms to be flagged as statistical anomaly")
	}
}
