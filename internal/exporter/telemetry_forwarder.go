package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type DatadogMetricSeries struct {
	Metric string      `json:"metric"`
	Points [][]float64 `json:"points"`
	Type   string      `json:"type"`
	Tags   []string    `json:"tags"`
}

type DatadogPayload struct {
	Series []DatadogMetricSeries `json:"series"`
}

type TelemetryForwarder struct {
	datadogAPIKey string
	client        *http.Client
}

func NewTelemetryForwarder(apiKey string) *TelemetryForwarder {
	return &TelemetryForwarder{
		datadogAPIKey: apiKey,
		client:        &http.Client{Timeout: 5 * time.Second},
	}
}

func (f *TelemetryForwarder) FormatOpenMetrics(result *models.CheckResult) string {
	val := 0.0
	if result.Status == models.StatusUp {
		val = 1.0
	}
	latencySec := result.Latency.Seconds()

	return fmt.Sprintf(
		"# TYPE pulsegrid_probe_success gauge
"+
			"pulsegrid_probe_success{target_id="%s"} %.1f
"+
			"# TYPE pulsegrid_probe_duration_seconds gauge
"+
			"pulsegrid_probe_duration_seconds{target_id="%s"} %.6f
",
		result.TargetID, val,
		result.TargetID, latencySec,
	)
}

func (f *TelemetryForwarder) ForwardToDatadog(ctx context.Context, result *models.CheckResult) error {
	if f.datadogAPIKey == "" {
		return nil // skip if not configured
	}

	val := 0.0
	if result.Status == models.StatusUp {
		val = 1.0
	}
	now := float64(time.Now().Unix())

	payload := DatadogPayload{
		Series: []DatadogMetricSeries{
			{
				Metric: "pulsegrid.probe.success",
				Points: [][]float64{{now, val}},
				Type:   "gauge",
				Tags:   []string{fmt.Sprintf("target_id:%s", result.TargetID)},
			},
			{
				Metric: "pulsegrid.probe.latency_ms",
				Points: [][]float64{{now, float64(result.Latency.Milliseconds())}},
				Type:   "gauge",
				Tags:   []string{fmt.Sprintf("target_id:%s", result.TargetID)},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := "https://api.datadoghq.com/api/v1/series"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("DD-API-KEY", f.datadogAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("datadog api error: %s", resp.Status)
	}
	return nil
}
