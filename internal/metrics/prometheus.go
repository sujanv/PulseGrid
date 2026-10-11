package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sujanv/PulseGrid/internal/models"
)

var (
	ChecksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pulsegrid_checks_total",
			Help: "Total number of health checks executed by PulseGrid",
		},
		[]string{"target_id", "target_name", "type", "status"},
	)

	CheckDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pulsegrid_check_duration_seconds",
			Help:    "Total health check execution duration in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"target_id", "type"},
	)

	DNSDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pulsegrid_dns_duration_seconds",
			Help:    "DNS lookup duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"target_id"},
	)

	TLSDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pulsegrid_tls_duration_seconds",
			Help:    "TLS handshake duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"target_id"},
	)

	TTFBDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pulsegrid_ttfb_duration_seconds",
			Help:    "Time to first byte (TTFB) in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"target_id"},
	)

	TargetUp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pulsegrid_target_up",
			Help: "Current operational status of target (1 = UP, 0 = DOWN)",
		},
		[]string{"target_id", "target_name"},
	)

	ActiveWorkers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "pulsegrid_active_workers",
			Help: "Number of worker goroutines currently executing checks",
		},
	)

	QueuedJobs = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "pulsegrid_queued_jobs",
			Help: "Number of check jobs pending in worker pool queue",
		},
	)

	ActiveIncidents = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "pulsegrid_active_incidents",
			Help: "Current count of active downtime incidents",
		},
	)

	AlertsDispatchedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pulsegrid_alerts_dispatched_total",
			Help: "Total alerts sent across channels",
		},
		[]string{"channel_type", "event_type"},
	)

	LatencyAnomaliesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pulsegrid_latency_anomalies_total",
			Help: "Total count of statistical latency spikes detected",
		},
		[]string{"target_id"},
	)
)

func RecordResult(target *models.Target, res *models.CheckResult) {
	statusStr := string(res.Status)
	ChecksTotal.WithLabelValues(target.ID, target.Name, string(target.Type), statusStr).Inc()
	CheckDuration.WithLabelValues(target.ID, string(target.Type)).Observe(res.Latency.Seconds())

	if res.Breakdown.DNSLookup > 0 {
		DNSDuration.WithLabelValues(target.ID).Observe(res.Breakdown.DNSLookup.Seconds())
	}
	if res.Breakdown.TLSHandshake > 0 {
		TLSDuration.WithLabelValues(target.ID).Observe(res.Breakdown.TLSHandshake.Seconds())
	}
	if res.Breakdown.TTFB > 0 {
		TTFBDuration.WithLabelValues(target.ID).Observe(res.Breakdown.TTFB.Seconds())
	}

	upVal := 0.0
	if res.Status == models.StatusUp {
		upVal = 1.0
	}
	TargetUp.WithLabelValues(target.ID, target.Name).Set(upVal)
}
