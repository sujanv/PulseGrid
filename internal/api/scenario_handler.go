package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/chaos"
	"github.com/sujanv/PulseGrid/internal/integrations"
	"github.com/sujanv/PulseGrid/internal/rca"
	"github.com/sujanv/PulseGrid/internal/synthetic"
)

type ScenarioHandler struct {
	runner        *synthetic.ScenarioRunner
	scheduler     *chaos.ChaosScheduler
	postmortemGen *rca.PostMortemGenerator
	alertmanager  *integrations.AlertmanagerReceiver
}

func NewScenarioHandler(scheduler *chaos.ChaosScheduler) *ScenarioHandler {
	return &ScenarioHandler{
		runner:        synthetic.NewScenarioRunner(),
		scheduler:     scheduler,
		postmortemGen: rca.NewPostMortemGenerator(),
		alertmanager:  integrations.NewAlertmanagerReceiver(nil),
	}
}

func (h *ScenarioHandler) HandleRunScenario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name  string                   `json:"name"`
		Steps []synthetic.ScenarioStep `json:"steps"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	res, err := h.runner.Execute(r.Context(), uuid.New().String(), req.Name, req.Steps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *ScenarioHandler) HandleGeneratePostMortem(w http.ResponseWriter, r *http.Request) {
	incidentID := r.URL.Query().Get("incident_id")
	if incidentID == "" {
		incidentID = "INC-1001"
	}

	now := time.Now().UTC()
	data := &rca.IncidentData{
		IncidentID:     incidentID,
		TargetID:       "tgt-payment-gw",
		TargetName:     "Core Payment API",
		TargetURL:      "https://api.payment.internal",
		RootCause:      rca.CauseGateway502,
		Recommendation: "Scale up backend worker pods and inspect connection pool limits.",
		StartedAt:      now.Add(-25 * time.Minute),
		ResolvedAt:     now,
		AffectedRegions: []string{"us-east", "eu-west"},
		Timeline: []rca.IncidentTimelineEntry{
			{Timestamp: now.Add(-25 * time.Minute), Description: "Health probe failure: 502 Bad Gateway", Severity: "CRITICAL", Source: "PulseGrid-Probe"},
			{Timestamp: now.Add(-20 * time.Minute), Description: "Escalation policy triggered: Paged On-Call", Severity: "HIGH", Source: "PagerDuty"},
			{Timestamp: now.Add(-2 * time.Minute), Description: "Pod restart completed; HTTP 200 restored", Severity: "INFO", Source: "K8s-AutoHealer"},
		},
	}

	md := h.postmortemGen.GenerateMarkdown(data)
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	_, _ = w.Write([]byte(md))
}

func (h *ScenarioHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/scenarios/run", h.HandleRunScenario)
	mux.HandleFunc("/api/v1/incidents/postmortem", h.HandleGeneratePostMortem)
	mux.Handle("/api/v1/integrations/alertmanager", h.alertmanager)
}
