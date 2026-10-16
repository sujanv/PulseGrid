package network

import (
	"math"
	"time"
)

type PacketProbeSample struct {
	Sequence  int           `json:"sequence"`
	SendTime  time.Time     `json:"send_time"`
	RecvTime  time.Time     `json:"recv_time"`
	RTT       time.Duration `json:"rtt"`
	Lost      bool          `json:"lost"`
}

type JitterProfileResult struct {
	TargetHost      string        `json:"target_host"`
	PacketsSent     int           `json:"packets_sent"`
	PacketsReceived int           `json:"packets_received"`
	PacketLossPct   float64       `json:"packet_loss_pct"`
	MeanRTT         time.Duration `json:"mean_rtt"`
	MinRTT          time.Duration `json:"min_rtt"`
	MaxRTT          time.Duration `json:"max_rtt"`
	RFC1889Jitter   time.Duration `json:"rfc1889_jitter"` // smoothed inter-arrival jitter
	EvaluatedAt     time.Time     `json:"evaluated_at"`
}

type JitterProfiler struct{}

func NewJitterProfiler() *JitterProfiler {
	return &JitterProfiler{}
}

// ComputeJitter calculates RFC 1889 / RFC 3550 inter-arrival jitter:
// D(i, j) = (R_j - S_j) - (R_i - S_i)
// J = J + (|D(i, j)| - J) / 16
func (p *JitterProfiler) ComputeJitter(host string, samples []PacketProbeSample) JitterProfileResult {
	if len(samples) == 0 {
		return JitterProfileResult{TargetHost: host, EvaluatedAt: time.Now().UTC()}
	}

	var receivedCount int
	var totalRTT time.Duration
	minRTT := time.Duration(math.MaxInt64)
	maxRTT := time.Duration(0)

	var prevTransit time.Duration
	var jitter float64
	hasPrev := false

	for _, s := range samples {
		if s.Lost {
			continue
		}
		receivedCount++
		totalRTT += s.RTT

		if s.RTT < minRTT {
			minRTT = s.RTT
		}
		if s.RTT > maxRTT {
			maxRTT = s.RTT
		}

		transit := s.RecvTime.Sub(s.SendTime)
		if hasPrev {
			d := math.Abs(float64(transit - prevTransit))
			jitter = jitter + (d-jitter)/16.0
		}
		prevTransit = transit
		hasPrev = true
	}

	lossPct := 0.0
	meanRTT := time.Duration(0)
	if len(samples) > 0 {
		lossPct = float64(len(samples)-receivedCount) / float64(len(samples)) * 100.0
	}
	if receivedCount > 0 {
		meanRTT = totalRTT / time.Duration(receivedCount)
	} else {
		minRTT = 0
	}

	return JitterProfileResult{
		TargetHost:      host,
		PacketsSent:     len(samples),
		PacketsReceived: receivedCount,
		PacketLossPct:   lossPct,
		MeanRTT:         meanRTT,
		MinRTT:          minRTT,
		MaxRTT:          maxRTT,
		RFC1889Jitter:   time.Duration(jitter),
		EvaluatedAt:     time.Now().UTC(),
	}
}
