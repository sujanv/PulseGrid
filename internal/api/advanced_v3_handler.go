package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sujanv/PulseGrid/internal/canary"
	"github.com/sujanv/PulseGrid/internal/checker"
	"github.com/sujanv/PulseGrid/internal/synthetic"
)

type AdvancedV3Handler struct {
	harExporter     *synthetic.HARExporter
	canaryEvaluator *canary.CanaryEvaluator
	certInspector   *checker.CertInspector
}

func NewAdvancedV3Handler() *AdvancedV3Handler {
	return &AdvancedV3Handler{
		harExporter:     synthetic.NewHARExporter(),
		canaryEvaluator: canary.NewCanaryEvaluator(0.02, 50*time.Millisecond),
		certInspector:   checker.NewCertInspector(),
	}
}

func (h *AdvancedV3Handler) HandleExportHAR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dummyResult := &synthetic.ScenarioResult{
		ScenarioID:    "scen-demo",
		ScenarioName:  "User Checkout Journey",
		Success:       true,
		ExecutedAt:    time.Now().UTC(),
		TotalDuration: 250 * time.Millisecond,
		StepResults: []synthetic.ScenarioStepResult{
			{
				StepName:   "Login Step",
				URL:        "https://api.example.com/v1/auth",
				StatusCode: 200,
				Success:    true,
				Waterfall: synthetic.StepPerformanceWaterfall{
					DNSDuration:      5 * time.Millisecond,
					ConnectDuration:  10 * time.Millisecond,
					TLSDuration:      20 * time.Millisecond,
					TTFB:             60 * time.Millisecond,
					TransferDuration: 15 * time.Millisecond,
					TotalDuration:    110 * time.Millisecond,
				},
			},
		},
	}

	har := h.harExporter.ExportScenarioToHAR(dummyResult)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(har)
}

func (h *AdvancedV3Handler) HandleCanaryEvaluation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		BaselineID string                      `json:"baseline_id"`
		CanaryID   string                      `json:"canary_id"`
		Baseline   canary.TargetMetricsSummary `json:"baseline"`
		Canary     canary.TargetMetricsSummary `json:"canary"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	res := h.canaryEvaluator.Evaluate(req.BaselineID, req.CanaryID, req.Baseline, req.Canary)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *AdvancedV3Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/targets/har/export", h.HandleExportHAR)
	mux.HandleFunc("/api/v1/canary/evaluate", h.HandleCanaryEvaluation)
}
