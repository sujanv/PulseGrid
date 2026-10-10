package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sujanv/PulseGrid/internal/storage/postgres"
)

type ResultHandler struct {
	repo postgres.ResultRepository
}

func NewResultHandler(repo postgres.ResultRepository) *ResultHandler {
	return &ResultHandler{repo: repo}
}

func (h *ResultHandler) HandleResults(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		ErrorResponse(w, http.StatusBadRequest, "invalid results path")
		return
	}
	targetID := parts[3]

	limit := 50
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := h.repo.GetRecentByTarget(r.Context(), targetID, limit)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, results)
}

func (h *ResultHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		ErrorResponse(w, http.StatusBadRequest, "invalid stats path")
		return
	}
	targetID := parts[3]

	window := 24 * time.Hour
	stats, err := h.repo.GetLatencyStats(r.Context(), targetID, window)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	uptime, _ := h.repo.CalculateUptime(r.Context(), targetID, window)

	response := map[string]interface{}{
		"target_id":            targetID,
		"window_hours":         24,
		"uptime_percentage":    uptime,
		"latency_stats":        stats,
	}
	JSONResponse(w, http.StatusOK, response)
}
