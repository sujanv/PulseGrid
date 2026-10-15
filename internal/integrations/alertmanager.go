package integrations

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AlertmanagerAlert struct {
	Status       string            `json:"status"` // "firing" or "resolved"
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     time.Time         `json:"startsAt"`
	EndsAt       time.Time         `json:"endsAt,omitempty"`
	GeneratorURL string            `json:"generatorURL"`
	Fingerprint  string            `json:"fingerprint"`
}

type AlertmanagerWebhookPayload struct {
	Receiver          string              `json:"receiver"`
	Status            string              `json:"status"`
	Alerts            []AlertmanagerAlert `json:"alerts"`
	GroupLabels       map[string]string   `json:"groupLabels"`
	CommonLabels      map[string]string   `json:"commonLabels"`
	CommonAnnotations map[string]string   `json:"commonAnnotations"`
	ExternalURL       string              `json:"externalURL"`
}

type AlertmanagerReceiver struct {
	OnAlertIngested func(alert *AlertmanagerAlert)
}

func NewAlertmanagerReceiver(onAlert func(alert *AlertmanagerAlert)) *AlertmanagerReceiver {
	return &AlertmanagerReceiver{OnAlertIngested: onAlert}
}

func (r *AlertmanagerReceiver) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	var payload AlertmanagerWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, fmt.Sprintf("invalid json payload: %v", err), http.StatusBadRequest)
		return
	}

	if r.OnAlertIngested != nil {
		for i := range payload.Alerts {
			r.OnAlertIngested(&payload.Alerts[i])
		}
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"success"}`))
}
