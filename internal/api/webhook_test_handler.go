package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/alert"
	"github.com/sujanv/PulseGrid/internal/models"
)

func HandleWebhookTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		ErrorResponse(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		URL       string `json:"url"`
		SecretKey string `json:"secret_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		ErrorResponse(w, http.StatusBadRequest, "valid webhook url required")
		return
	}

	channel := alert.NewWebhookChannel(req.URL, req.SecretKey)
	testEvent := &models.AlertEvent{
		ID:         uuid.New().String(),
		TargetID:   "test-target-id",
		TargetName: "Test Target Endpoint",
		TargetURL:  "https://example.com/healthz",
		EventType:  models.AlertOutage,
		Message:    "Simulated test alert delivery from PulseGrid",
		LatencyMs:  42.5,
		StatusCode: 500,
		Timestamp:  time.Now().UTC(),
	}

	err := channel.Send(r.Context(), testEvent)
	if err != nil {
		JSONResponse(w, http.StatusBadGateway, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Webhook test event delivered successfully",
	})
}
