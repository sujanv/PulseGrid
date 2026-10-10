package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sujanv/PulseGrid/internal/maintenance"
)

type MaintenanceHandler struct {
	manager *maintenance.Manager
}

func NewMaintenanceHandler(m *maintenance.Manager) *MaintenanceHandler {
	return &MaintenanceHandler{manager: m}
}

func (h *MaintenanceHandler) HandleWindows(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list := h.manager.ListWindows()
		JSONResponse(w, http.StatusOK, list)
	case http.MethodPost:
		var req struct {
			Title       string    `json:"title"`
			Description string    `json:"description"`
			TargetIDs   []string  `json:"target_ids"`
			StartTime   time.Time `json:"start_time"`
			EndTime     time.Time `json:"end_time"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			ErrorResponse(w, http.StatusBadRequest, "invalid json body")
			return
		}
		if req.Title == "" || req.EndTime.Before(req.StartTime) {
			ErrorResponse(w, http.StatusBadRequest, "valid title and time range required")
			return
		}
		win := h.manager.AddWindow(req.Title, req.Description, req.TargetIDs, req.StartTime, req.EndTime)
		JSONResponse(w, http.StatusCreated, win)
	case http.MethodDelete:
		id := ExtractPathID(r.URL.Path, "/api/v1/maintenance/")
		if id == "" {
			ErrorResponse(w, http.StatusBadRequest, "missing window id")
			return
		}
		if h.manager.RemoveWindow(id) {
			JSONResponse(w, http.StatusOK, map[string]string{"message": "maintenance window deleted"})
		} else {
			ErrorResponse(w, http.StatusNotFound, "window not found")
		}
	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
