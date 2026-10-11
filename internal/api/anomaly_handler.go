package api

import (
	"net/http"
	"strings"

	"github.com/sujanv/PulseGrid/internal/anomaly"
)

type AnomalyHandler struct {
	detector *anomaly.AnomalyDetector
}

func NewAnomalyHandler(d *anomaly.AnomalyDetector) *AnomalyHandler {
	return &AnomalyHandler{detector: d}
}

func (h *AnomalyHandler) HandleAnomalies(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		ErrorResponse(w, http.StatusBadRequest, "invalid path")
		return
	}
	targetID := parts[3]

	history := h.detector.GetHistory(targetID)
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"target_id": targetID,
		"anomalies": history,
	})
}
