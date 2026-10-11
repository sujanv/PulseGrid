package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sujanv/PulseGrid/internal/storage/postgres"
)

type StatusPageHandler struct {
	targetRepo   postgres.TargetRepository
	resultRepo   postgres.ResultRepository
	incidentRepo postgres.IncidentRepository
}

func NewStatusPageHandler(t postgres.TargetRepository, r postgres.ResultRepository, i postgres.IncidentRepository) *StatusPageHandler {
	return &StatusPageHandler{
		targetRepo:   t,
		resultRepo:   r,
		incidentRepo: i,
	}
}

func (h *StatusPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	activeIncidents := 0
	if h.incidentRepo != nil {
		incs, err := h.incidentRepo.ListActiveIncidents(ctx)
		if err == nil {
			activeIncidents = len(incs)
		}
	}

	bannerText := "All Systems Operational"
	bannerColor := "#238636"
	if activeIncidents > 0 {
		bannerText = fmt.Sprintf("Service Degradation (%d active incident(s))", activeIncidents)
		bannerColor = "#da3633"
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>PulseGrid Status Page</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0d1117; color: #c9d1d9; padding: 40px 20px; }
    .container { max-width: 900px; margin: 0 auto; }
    .banner { background: %s; color: white; padding: 18px 24px; border-radius: 8px; font-weight: 700; font-size: 18px; margin-bottom: 32px; display: flex; align-items: center; gap: 12px; }
    .section-title { font-size: 20px; font-weight: 600; color: #fff; margin-bottom: 16px; border-bottom: 1px solid #30363d; padding-bottom: 8px; }
    .card { background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 20px; margin-bottom: 16px; }
    .footer { text-align: center; color: #8b949e; font-size: 13px; margin-top: 48px; }
  </style>
</head>
<body>
  <div class="container">
    <h1 style="color: #fff; margin-bottom: 24px;">PulseGrid System Status</h1>
    <div class="banner">
      <span>●</span> %s
    </div>

    <div class="section-title">Core Services</div>
    <div class="card" style="display: flex; justify-content: space-between; align-items: center;">
      <div>
        <strong>API Gateway & Probes</strong>
        <div style="color: #8b949e; font-size: 13px; margin-top: 4px;">Bounded concurrency worker execution</div>
      </div>
      <span style="color: #3fb950; font-weight: 600;">Operational</span>
    </div>

    <div class="card" style="display: flex; justify-content: space-between; align-items: center;">
      <div>
        <strong>Multi-Protocol Schedulers</strong>
        <div style="color: #8b949e; font-size: 13px; margin-top: 4px;">HTTP, TCP, DNS, TLS, Ping, WebSocket, gRPC</div>
      </div>
      <span style="color: #3fb950; font-weight: 600;">Operational</span>
    </div>

    <div class="card" style="display: flex; justify-content: space-between; align-items: center;">
      <div>
        <strong>Alert Notification Dispatcher</strong>
        <div style="color: #8b949e; font-size: 13px; margin-top: 4px;">Webhooks, Slack, Discord, Telegram, PagerDuty, Teams</div>
      </div>
      <span style="color: #3fb950; font-weight: 600;">Operational</span>
    </div>

    <div class="footer">
      Powered by PulseGrid Uptime Monitoring • Updated %s
    </div>
  </div>
</body>
</html>`, bannerColor, bannerText, time.Now().UTC().Format(time.RFC1123))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}
