package synthetic

import (
	"encoding/json"
	"time"
)

type HARNameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HARTimings struct {
	Blocked int64 `json:"blocked"`
	DNS     int64 `json:"dns"`
	Connect int64 `json:"connect"`
	Send    int64 `json:"send"`
	Wait    int64 `json:"wait"`
	Receive int64 `json:"receive"`
	SSL     int64 `json:"ssl"`
}

type HARRequest struct {
	Method      string         `json:"method"`
	URL         string         `json:"url"`
	HTTPVersion string         `json:"httpVersion"`
	Headers     []HARNameValue `json:"headers"`
	BodySize    int64          `json:"bodySize"`
}

type HARContent struct {
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
}

type HARResponse struct {
	Status      int            `json:"status"`
	StatusText  string         `json:"statusText"`
	HTTPVersion string         `json:"httpVersion"`
	Headers     []HARNameValue `json:"headers"`
	Content     HARContent     `json:"content"`
	RedirectURL string         `json:"redirectURL"`
	BodySize    int64          `json:"bodySize"`
}

type HAREntry struct {
	StartedDateTime time.Time   `json:"startedDateTime"`
	Time            int64       `json:"time"` // total ms
	Request         HARRequest  `json:"request"`
	Response        HARResponse `json:"response"`
	Timings         HARTimings  `json:"timings"`
}

type HARLog struct {
	Version string     `json:"version"`
	Creator map[string]string `json:"creator"`
	Entries []HAREntry `json:"entries"`
}

type HARArchive struct {
	Log HARLog `json:"log"`
}

type HARExporter struct{}

func NewHARExporter() *HARExporter {
	return &HARExporter{}
}

func (e *HARExporter) ExportScenarioToHAR(result *ScenarioResult) *HARArchive {
	entries := make([]HAREntry, 0, len(result.StepResults))
	stepTime := result.ExecutedAt

	for _, step := range result.StepResults {
		totalMs := step.Waterfall.TotalDuration.Milliseconds()
		dnsMs := step.Waterfall.DNSDuration.Milliseconds()
		connMs := step.Waterfall.ConnectDuration.Milliseconds()
		tlsMs := step.Waterfall.TLSDuration.Milliseconds()
		ttfbMs := step.Waterfall.TTFB.Milliseconds()
		transferMs := step.Waterfall.TransferDuration.Milliseconds()

		statusText := "OK"
		if !step.Success {
			statusText = "Error"
		}

		entry := HAREntry{
			StartedDateTime: stepTime,
			Time:            totalMs,
			Request: HARRequest{
				Method:      "GET",
				URL:         step.URL,
				HTTPVersion: "HTTP/1.1",
				Headers: []HARNameValue{
					{Name: "User-Agent", Value: "PulseGrid-Synthetic/2.5"},
				},
				BodySize: 0,
			},
			Response: HARResponse{
				Status:      step.StatusCode,
				StatusText:  statusText,
				HTTPVersion: "HTTP/1.1",
				Headers: []HARNameValue{
					{Name: "Content-Type", Value: "text/html; charset=utf-8"},
				},
				Content: HARContent{
					Size:     1024,
					MimeType: "text/html",
				},
				BodySize: 1024,
			},
			Timings: HARTimings{
				DNS:     dnsMs,
				Connect: connMs,
				SSL:     tlsMs,
				Send:    1,
				Wait:    ttfbMs,
				Receive: transferMs,
			},
		}

		entries = append(entries, entry)
		stepTime = stepTime.Add(step.Waterfall.TotalDuration)
	}

	return &HARArchive{
		Log: HARLog{
			Version: "1.2",
			Creator: map[string]string{
				"name":    "PulseGrid Enterprise",
				"version": "2.5.0",
			},
			Entries: entries,
		},
	}
}

func (e *HARExporter) ToJSON(archive *HARArchive) (string, error) {
	bytes, err := json.MarshalIndent(archive, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
