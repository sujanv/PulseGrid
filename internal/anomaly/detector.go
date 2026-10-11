package anomaly

import (
	"sync"
	"time"
)

type AnomalyEvent struct {
	TargetID    string    `json:"target_id"`
	Timestamp   time.Time `json:"timestamp"`
	ObservedMs  float64   `json:"observed_ms"`
	BaselineMs  float64   `json:"baseline_ms"`
	StdDevMs    float64   `json:"std_dev_ms"`
	ZScore      float64   `json:"z_score"`
	Description string    `json:"description"`
}

type AnomalyDetector struct {
	mu          sync.RWMutex
	targetStats map[string]*WelfordStats
	history     map[string][]AnomalyEvent
	zThreshold  float64
}

func NewAnomalyDetector(zThreshold float64) *AnomalyDetector {
	if zThreshold <= 0 {
		zThreshold = 3.0 // 3 standard deviations = 99.7% statistical threshold
	}
	return &AnomalyDetector{
		targetStats: make(map[string]*WelfordStats),
		history:     make(map[string][]AnomalyEvent),
		zThreshold:  zThreshold,
	}
}

func (d *AnomalyDetector) Observe(targetID string, latencyMs float64) (*AnomalyEvent, bool) {
	d.mu.Lock()
	stats, ok := d.targetStats[targetID]
	if !ok {
		stats = NewWelfordStats()
		d.targetStats[targetID] = stats
	}
	d.mu.Unlock()

	// Need minimum observations before flagging anomalies
	if stats.Count >= 20 {
		z := stats.ZScore(latencyMs)
		if z > d.zThreshold {
			ev := AnomalyEvent{
				TargetID:    targetID,
				Timestamp:   time.Now().UTC(),
				ObservedMs:  latencyMs,
				BaselineMs:  stats.Mean,
				StdDevMs:    stats.StdDev(),
				ZScore:      z,
				Description: "Statistical latency anomaly detected: Z-Score exceeds threshold",
			}

			d.mu.Lock()
			d.history[targetID] = append(d.history[targetID], ev)
			d.mu.Unlock()

			// Update stats after check
			stats.Update(latencyMs)
			return &ev, true
		}
	}

	stats.Update(latencyMs)
	return nil, false
}

func (d *AnomalyDetector) GetHistory(targetID string) []AnomalyEvent {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.history[targetID]
}
