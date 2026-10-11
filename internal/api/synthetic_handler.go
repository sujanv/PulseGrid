package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/storage/postgres"
	"github.com/sujanv/PulseGrid/internal/synthetic"
)

type SyntheticHandler struct {
	targetRepo postgres.TargetRepository
	runner     *synthetic.Runner
}

func NewSyntheticHandler(repo postgres.TargetRepository) *SyntheticHandler {
	return &SyntheticHandler{
		targetRepo: repo,
		runner:     synthetic.NewRunner(30 * time.Second),
	}
}

func (h *SyntheticHandler) HandleCreateSynthetic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Name     string             `json:"name"`
		Interval int                `json:"interval_seconds"`
		Workflow synthetic.Workflow `json:"workflow"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if len(req.Workflow.Steps) == 0 {
		ErrorResponse(w, http.StatusBadRequest, "at least one workflow step is required")
		return
	}

	wfBytes, _ := json.Marshal(req.Workflow)
	interval := time.Duration(req.Interval) * time.Second
	if interval < 10*time.Second {
		interval = 60 * time.Second
	}

	target := &models.Target{
		ID:               uuid.New().String(),
		Name:             req.Name,
		URL:              req.Workflow.Steps[0].URL,
		Type:             models.TargetSynthetic,
		Interval:         interval,
		Timeout:          30 * time.Second,
		WorkflowJSON:     string(wfBytes),
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Active:           true,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	if h.targetRepo != nil {
		if err := h.targetRepo.Create(r.Context(), target); err != nil {
			ErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	JSONResponse(w, http.StatusCreated, target)
}

func (h *SyntheticHandler) HandleTestRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var wf synthetic.Workflow
	if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid json workflow")
		return
	}

	_, execResult := h.runner.RunWorkflow(r.Context(), &wf)
	JSONResponse(w, http.StatusOK, execResult)
}
