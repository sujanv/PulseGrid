package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/sujanv/PulseGrid/internal/network"
	"github.com/sujanv/PulseGrid/internal/rca"
	"github.com/sujanv/PulseGrid/internal/slo"
	"github.com/sujanv/PulseGrid/internal/statuspage"
)

type EnterpriseHandler struct {
	tracerouteEngine *network.TracerouteEngine
	burnEvaluator    *slo.BurnRateEvaluator
	budgetCalculator *slo.ErrorBudgetCalculator
	statusManager    *statuspage.StatusPageManager
	feedGen          *statuspage.FeedGenerator
	rcaEngine        *rca.RCAEngine
}

func NewEnterpriseHandler() *EnterpriseHandler {
	return &EnterpriseHandler{
		tracerouteEngine: network.NewTracerouteEngine(20, 2*time.Second),
		burnEvaluator:    slo.NewBurnRateEvaluator(nil),
		budgetCalculator: slo.NewErrorBudgetCalculator(),
		statusManager:    statuspage.NewStatusPageManager(),
		feedGen:          statuspage.NewFeedGenerator("PulseGrid Status", "https://pulsegrid.io"),
		rcaEngine:        rca.NewRCAEngine(),
	}
}

func (h *EnterpriseHandler) HandleTraceroute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	host := r.URL.Query().Get("host")
	if host == "" {
		http.Error(w, "host query parameter is required", http.StatusBadRequest)
		return
	}

	res, err := h.tracerouteEngine.Trace(r.Context(), host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *EnterpriseHandler) HandleSLOBurnRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	targetID := r.URL.Query().Get("target_id")
	sloStr := r.URL.Query().Get("slo")
	sloObj := 99.9
	if v, err := strconv.ParseFloat(sloStr, 64); err == nil && v > 0 {
		sloObj = v
	}

	res := h.burnEvaluator.Evaluate(targetID, sloObj, 1*time.Hour, 10000, 25)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (h *EnterpriseHandler) HandleStatusComponents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	components := h.statusManager.ListComponents()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"overall_status": h.statusManager.GetOverallStatus(),
		"components":     components,
	})
}

func (h *EnterpriseHandler) HandleRSSFeed(w http.ResponseWriter, r *http.Request) {
	feed := h.feedGen.GenerateRSS([]statuspage.IncidentFeedItem{
		{
			ID:          "inc-1",
			Title:       "Degraded Latency in EU-West Region",
			Description: "Intermediate transit provider experiencing packet loss; traffic rerouted.",
			PubDate:     time.Now().UTC().Add(-2 * time.Hour),
			Link:        "https://pulsegrid.io/incidents/inc-1",
		},
	})
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	_, _ = w.Write([]byte(feed))
}

func (h *EnterpriseHandler) HandleAtomFeed(w http.ResponseWriter, r *http.Request) {
	feed := h.feedGen.GenerateAtom([]statuspage.IncidentFeedItem{
		{
			ID:          "inc-1",
			Title:       "Degraded Latency in EU-West Region",
			Description: "Intermediate transit provider experiencing packet loss; traffic rerouted.",
			PubDate:     time.Now().UTC().Add(-2 * time.Hour),
			Link:        "https://pulsegrid.io/incidents/inc-1",
		},
	})
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	_, _ = w.Write([]byte(feed))
}

func (h *EnterpriseHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/network/traceroute", h.HandleTraceroute)
	mux.HandleFunc("/api/v1/slo/burn-rate", h.HandleSLOBurnRate)
	mux.HandleFunc("/api/v1/statuspage/components", h.HandleStatusComponents)
	mux.HandleFunc("/status/feed.rss", h.HandleRSSFeed)
	mux.HandleFunc("/status/feed.atom", h.HandleAtomFeed)
}
