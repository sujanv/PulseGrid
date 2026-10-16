package exporter

import (
	"fmt"
	"strings"

	"github.com/sujanv/PulseGrid/internal/models"
)

type RemoteWriteTimeSeries struct {
	Labels    map[string]string `json:"labels"`
	Timestamp int64             `json:"timestamp"` // ms epoch
	Value     float64           `json:"value"`
}

type RemoteWriteClient struct {
	endpoint string
}

func NewRemoteWriteClient(endpoint string) *RemoteWriteClient {
	if endpoint == "" {
		endpoint = "http://localhost:9090/api/v1/write"
	}
	return &RemoteWriteClient{endpoint: endpoint}
}

func (c *RemoteWriteClient) BuildTimeSeries(result *models.CheckResult) []RemoteWriteTimeSeries {
	ts := result.Timestamp.UnixMilli()
	statusVal := 0.0
	if result.Status == models.StatusUp {
		statusVal = 1.0
	}

	return []RemoteWriteTimeSeries{
		{
			Labels: map[string]string{
				"__name__":  "pulsegrid_target_up",
				"target_id": result.TargetID,
			},
			Timestamp: ts,
			Value:     statusVal,
		},
		{
			Labels: map[string]string{
				"__name__":  "pulsegrid_target_latency_seconds",
				"target_id": result.TargetID,
			},
			Timestamp: ts,
			Value:     result.Latency.Seconds(),
		},
	}
}

func (c *RemoteWriteClient) FormatDebugPayload(series []RemoteWriteTimeSeries) string {
	var sb strings.Builder
	for _, s := range series {
		sb.WriteString(fmt.Sprintf("%s{target_id=\"%s\"} %f @ %d\n", s.Labels["__name__"], s.Labels["target_id"], s.Value, s.Timestamp))
	}
	return sb.String()
}
