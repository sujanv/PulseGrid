package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sujanv/PulseGrid/internal/alert"
	"github.com/sujanv/PulseGrid/internal/integrations"
	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/network"
	"github.com/sujanv/PulseGrid/internal/synthetic"
	"github.com/sujanv/PulseGrid/internal/web"
)

type AdvancedV4Handler struct {
	vitalsCalc    *synthetic.WebVitalsCalculator
	jitterProf    *network.JitterProfiler
	badgeGen      *web.BadgeGenerator
	onCallManager *alert.OnCallRotationManager
	grafanaGen    *integrations.GrafanaDashboardGenerator
}

func NewAdvancedV4Handler() *AdvancedV4Handler {
	return &AdvancedV4Handler{
		vitalsCalc:    synthetic.NewWebVitalsCalculator(),
		jitterProf:    network.NewJitterProfiler(),
		badgeGen:      web.NewBadgeGenerator(),
		onCallManager: alert.NewOnCallRotationManager(),
		grafanaGen:    integrations.NewGrafanaDashboardGenerator(),
	}
}

func (h *AdvancedV4Handler) HandleWebVitals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL   string  `json:"url"`
		LCPMs int64   `json:"lcp_ms"`
		FIDMs int64   `json:"fid_ms"`
		CLS   float64 `json:"cls"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	rep := h.vitalsCalc.Calculate(req.URL, time.Duration(req.LCPMs)*time.Millisecond, time.Duration(req.FIDMs)*time.Millisecond, 200*time.Millisecond, req.CLS)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rep)
}

func (h *AdvancedV4Handler) HandleLiveBadge(w http.ResponseWriter, r *http.Request) {
	svg := h.badgeGen.GenerateSVG(models.StatusUp, 42*time.Millisecond, 99.98)
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	_, _ = w.Write([]byte(svg))
}

func (h *AdvancedV4Handler) HandleGrafanaExport(w http.ResponseWriter, r *http.Request) {
	jsonStr, err := h.grafanaGen.ExportDashboardJSON()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(jsonStr))
}

func (h *AdvancedV4Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/synthetic/web-vitals", h.HandleWebVitals)
	mux.HandleFunc("/api/v1/integrations/grafana/dashboard", h.HandleGrafanaExport)
	mux.HandleFunc("/badge/live/badge.svg", h.HandleLiveBadge)
}
