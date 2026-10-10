package api

import (
	"net/http"
	"strconv"

	"github.com/sujanv/PulseGrid/internal/storage/postgres"
)

type IncidentHandler struct {
	repo postgres.IncidentRepository
}

func NewIncidentHandler(repo postgres.IncidentRepository) *IncidentHandler {
	return &IncidentHandler{repo: repo}
}

func (h *IncidentHandler) HandleIncidents(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	incidents, err := h.repo.ListIncidents(r.Context(), limit, offset)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, incidents)
}

func (h *IncidentHandler) HandleActiveIncidents(w http.ResponseWriter, r *http.Request) {
	incidents, err := h.repo.ListActiveIncidents(r.Context())
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, incidents)
}
