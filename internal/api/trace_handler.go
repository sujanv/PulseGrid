package api

import (
	"net/http"
	"strings"

	"github.com/sujanv/PulseGrid/internal/tracing"
)

type TraceHandler struct {
	tracer *tracing.Tracer
}

func NewTraceHandler(t *tracing.Tracer) *TraceHandler {
	return &TraceHandler{tracer: t}
}

func (h *TraceHandler) HandleSpans(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		ErrorResponse(w, http.StatusBadRequest, "invalid traces path")
		return
	}
	targetID := parts[3]

	spans := h.tracer.GetSpans(targetID)
	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"target_id": targetID,
		"spans":     spans,
	})
}
