package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/remediation"
)

type RemediationHandler struct {
	rules  map[string]*remediation.RemediationRule
	logger *remediation.AuditLogger
}

func NewRemediationHandler(l *remediation.AuditLogger) *RemediationHandler {
	return &RemediationHandler{
		rules:  make(map[string]*remediation.RemediationRule),
		logger: l,
	}
}

func (h *RemediationHandler) HandleRemediation(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		ErrorResponse(w, http.StatusBadRequest, "invalid remediation path")
		return
	}
	targetID := parts[3]

	switch r.Method {
	case http.MethodGet:
		rule, ok := h.rules[targetID]
		logs := h.logger.GetLogs(targetID)
		JSONResponse(w, http.StatusOK, map[string]interface{}{
			"target_id": targetID,
			"rule":      rule,
			"configured": ok,
			"history":   logs,
		})
	case http.MethodPost:
		var req struct {
			URL             string `json:"url"`
			CooldownMinutes int    `json:"cooldown_minutes"`
			Enabled         bool   `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
			ErrorResponse(w, http.StatusBadRequest, "valid webhook url required")
			return
		}

		cooldown := time.Duration(req.CooldownMinutes) * time.Minute
		if cooldown < 1*time.Minute {
			cooldown = 15 * time.Minute
		}

		rule := &remediation.RemediationRule{
			ID:       uuid.New().String(),
			TargetID: targetID,
			Action:   remediation.ActionWebhook,
			URL:      req.URL,
			Cooldown: cooldown,
			Enabled:  req.Enabled,
		}
		h.rules[targetID] = rule
		JSONResponse(w, http.StatusCreated, rule)
	default:
		ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
