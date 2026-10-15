package network

import (
	"sync"
	"time"
)

type AnycastRegionMetric struct {
	Region     string        `json:"region"`
	ResolvedIP string        `json:"resolved_ip"`
	MeanRTT    time.Duration `json:"mean_rtt"`
	Variance   time.Duration `json:"variance"`
}

type AnycastShiftEvent struct {
	Host        string    `json:"host"`
	Region      string    `json:"region"`
	OldIP       string    `json:"old_ip"`
	NewIP       string    `json:"new_ip"`
	RTTDelta    time.Duration `json:"rtt_delta"`
	DetectedAt  time.Time `json:"detected_at"`
	Description string    `json:"description"`
}

type AnycastMatrixEvaluator struct {
	mu           sync.RWMutex
	lastObserved map[string]map[string]AnycastRegionMetric // host -> region -> metric
}

func NewAnycastMatrixEvaluator() *AnycastMatrixEvaluator {
	return &AnycastMatrixEvaluator{
		lastObserved: make(map[string]map[string]AnycastRegionMetric),
	}
}

func (e *AnycastMatrixEvaluator) RecordAndCheckShifts(host string, metrics []AnycastRegionMetric) []AnycastShiftEvent {
	e.mu.Lock()
	defer e.mu.Unlock()

	var shifts []AnycastShiftEvent
	regMap, exists := e.lastObserved[host]
	if !exists {
		regMap = make(map[string]AnycastRegionMetric)
		e.lastObserved[host] = regMap
	}

	now := time.Now().UTC()
	for _, m := range metrics {
		prev, ok := regMap[m.Region]
		if ok {
			// Check if Anycast BGP edge IP shifted
			if prev.ResolvedIP != m.ResolvedIP {
				rttDelta := m.MeanRTT - prev.MeanRTT
				shifts = append(shifts, AnycastShiftEvent{
					Host:        host,
					Region:      m.Region,
					OldIP:       prev.ResolvedIP,
					NewIP:       m.ResolvedIP,
					RTTDelta:    rttDelta,
					DetectedAt:  now,
					Description: "BGP anycast route shift detected: transit rerouted to new POP IP",
				})
			}
		}
		regMap[m.Region] = m
	}

	return shifts
}
