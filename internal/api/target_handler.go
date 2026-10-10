package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/storage/postgres"
)

type TargetHandler struct {
	repo postgres.TargetRepository
}

func NewTargetHandler(repo postgres.TargetRepository) *TargetHandler {
	return &TargetHandler{repo: repo}
}

func (h *TargetHandler) HandleTargets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTargets(w, r)
	case http.MethodPost:
		h.createTarget(w, r)
	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *TargetHandler) HandleTargetByID(w http.ResponseWriter, r *http.Request) {
	id := ExtractPathID(r.URL.Path, "/api/v1/targets/")
	if id == "" {
		ErrorResponse(w, http.StatusBadRequest, "missing target ID")
		return
	}

	// Handle sub-resources
	if strings.HasSuffix(r.URL.Path, "/check") && r.Method == http.MethodPost {
		JSONResponse(w, http.StatusOK, map[string]string{"status": "check queued", "target_id": id})
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getTarget(w, r, id)
	case http.MethodPut:
		h.updateTarget(w, r, id)
	case http.MethodDelete:
		h.deleteTarget(w, r, id)
	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *TargetHandler) listTargets(w http.ResponseWriter, r *http.Request) {
	targets, err := h.repo.ListAll(r.Context())
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, targets)
}

func (h *TargetHandler) getTarget(w http.ResponseWriter, r *http.Request, id string) {
	target, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		ErrorResponse(w, http.StatusNotFound, "target not found")
		return
	}
	JSONResponse(w, http.StatusOK, target)
}

func (h *TargetHandler) createTarget(w http.ResponseWriter, r *http.Request) {
	var target models.Target
	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := target.Validate(); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	target.ID = uuid.New().String()
	now := time.Now().UTC()
	target.CreatedAt = now
	target.UpdatedAt = now

	if err := h.repo.Create(r.Context(), &target); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONResponse(w, http.StatusCreated, target)
}

func (h *TargetHandler) updateTarget(w http.ResponseWriter, r *http.Request, id string) {
	var target models.Target
	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	target.ID = id
	if err := target.Validate(); err != nil {
		ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.repo.Update(r.Context(), &target); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSONResponse(w, http.StatusOK, target)
}

func (h *TargetHandler) deleteTarget(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.repo.Delete(r.Context(), id); err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"message": "target deleted"})
}
