package integrations

import (
	"encoding/json"
)

type GrafanaDashboardGenerator struct{}

func NewGrafanaDashboardGenerator() *GrafanaDashboardGenerator {
	return &GrafanaDashboardGenerator{}
}

func (g *GrafanaDashboardGenerator) ExportDashboardJSON() (string, error) {
	dashboard := map[string]interface{}{
		"title":       "PulseGrid Enterprise Monitoring",
		"timezone":    "browser",
		"schemaVersion": 38,
		"refresh":     "10s",
		"panels": []map[string]interface{}{
			{
				"id":    1,
				"title": "Overall Target Uptime (%)",
				"type":  "stat",
				"targets": []map[string]interface{}{
					{"expr": "avg(pulsegrid_target_up) * 100", "legendFormat": "Uptime %"},
				},
				"gridPos": map[string]int{"h": 6, "w": 6, "x": 0, "y": 0},
			},
			{
				"id":    2,
				"title": "Probe Latency (P50, P95, P99)",
				"type":  "timeseries",
				"targets": []map[string]interface{}{
					{"expr": "histogram_quantile(0.95, sum(rate(pulsegrid_probe_duration_seconds_bucket[5m])) by (le))", "legendFormat": "P95"},
					{"expr": "histogram_quantile(0.99, sum(rate(pulsegrid_probe_duration_seconds_bucket[5m])) by (le))", "legendFormat": "P99"},
				},
				"gridPos": map[string]int{"h": 8, "w": 18, "x": 6, "y": 0},
			},
			{
				"id":    3,
				"title": "Google SRE Error Budget Burn Rate",
				"type":  "gauge",
				"targets": []map[string]interface{}{
					{"expr": "pulsegrid_slo_burn_rate", "legendFormat": "{{target_id}}"},
				},
				"gridPos": map[string]int{"h": 8, "w": 8, "x": 0, "y": 6},
			},
		},
	}

	bytes, err := json.MarshalIndent(dashboard, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
