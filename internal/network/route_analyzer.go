package network

import (
	"fmt"
	"time"
)

type RouteAnomalyType string

const (
	AnomalyRouteFlap       RouteAnomalyType = "ROUTE_FLAP"
	AnomalyLatencySpike    RouteAnomalyType = "INTERMEDIATE_LATENCY_SPIKE"
	AnomalyPacketLoss      RouteAnomalyType = "PACKET_LOSS_BOTTLENECK"
)

type RouteAnomaly struct {
	Type        RouteAnomalyType `json:"type"`
	HopIndex    int              `json:"hop_index"`
	Description string           `json:"description"`
	Severity    string           `json:"severity"`
	DetectedAt  time.Time        `json:"detected_at"`
}

type RouteAnalyzer struct {
	latencyThreshold time.Duration
}

func NewRouteAnalyzer(latencyThreshold time.Duration) *RouteAnalyzer {
	if latencyThreshold <= 0 {
		latencyThreshold = 100 * time.Millisecond
	}
	return &RouteAnalyzer{latencyThreshold: latencyThreshold}
}

func (a *RouteAnalyzer) DetectAnomalies(prev, curr *TracerouteResult) []RouteAnomaly {
	var anomalies []RouteAnomaly
	if curr == nil {
		return anomalies
	}

	// 1. Detect Intermediate Latency Spikes
	for _, hop := range curr.Hops {
		if hop.RTT > a.latencyThreshold {
			anomalies = append(anomalies, RouteAnomaly{
				Type:        AnomalyLatencySpike,
				HopIndex:    hop.Index,
				Description: fmt.Sprintf("High latency at hop %d (%s): %v exceeds threshold %v", hop.Index, hop.IP, hop.RTT, a.latencyThreshold),
				Severity:    "WARNING",
				DetectedAt:  time.Now().UTC(),
			})
		}
		if hop.LossPercent > 10.0 {
			anomalies = append(anomalies, RouteAnomaly{
				Type:        AnomalyPacketLoss,
				HopIndex:    hop.Index,
				Description: fmt.Sprintf("Significant packet loss at hop %d (%s): %.1f%%", hop.Index, hop.IP, hop.LossPercent),
				Severity:    "CRITICAL",
				DetectedAt:  time.Now().UTC(),
			})
		}
	}

	// 2. Detect Route Flapping between consecutive traces
	if prev != nil {
		minHops := len(prev.Hops)
		if len(curr.Hops) < minHops {
			minHops = len(curr.Hops)
		}
		for i := 0; i < minHops; i++ {
			if prev.Hops[i].IP != curr.Hops[i].IP {
				anomalies = append(anomalies, RouteAnomaly{
					Type:        AnomalyRouteFlap,
					HopIndex:    i + 1,
					Description: fmt.Sprintf("Route flap at hop %d: shifted from %s to %s", i+1, prev.Hops[i].IP, curr.Hops[i].IP),
					Severity:    "INFO",
					DetectedAt:  time.Now().UTC(),
				})
			}
		}
	}

	return anomalies
}
