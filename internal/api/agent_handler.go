package api

import (
	"encoding/json"
	"net/http"

	"github.com/sujanv/PulseGrid/internal/agent"
)

type AgentHandler struct {
	coordinator *agent.Coordinator
}

func NewAgentHandler(c *agent.Coordinator) *AgentHandler {
	return &AgentHandler{coordinator: c}
}

func (h *AgentHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var info agent.AgentInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid agent info")
		return
	}

	h.coordinator.RegisterAgent(&info)
	JSONResponse(w, http.StatusOK, map[string]string{"status": "registered"})
}

func (h *AgentHandler) HandleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		AgentID    string `json:"agent_id"`
		ActiveJobs int    `json:"active_jobs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid json")
		return
	}

	if h.coordinator.Heartbeat(req.AgentID, req.ActiveJobs) {
		JSONResponse(w, http.StatusOK, map[string]string{"status": "alive"})
	} else {
		ErrorResponse(w, http.StatusNotFound, "agent not found")
	}
}

func (h *AgentHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	agents := h.coordinator.ActiveAgents()
	JSONResponse(w, http.StatusOK, agents)
}
