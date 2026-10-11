package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sujanv/PulseGrid/internal/storage/postgres"
)

type BadgeHandler struct {
	resultRepo postgres.ResultRepository
}

func NewBadgeHandler(r postgres.ResultRepository) *BadgeHandler {
	return &BadgeHandler{resultRepo: r}
}

func (b *BadgeHandler) ServeBadge(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		http.Error(w, "invalid badge url", http.StatusBadRequest)
		return
	}

	targetID := strings.TrimSuffix(parts[1], ".svg")
	uptime := 100.0
	if b.resultRepo != nil {
		up, err := b.resultRepo.CalculateUptime(r.Context(), targetID, 30*24*time.Hour)
		if err == nil {
			uptime = up
		}
	}

	color := "#4c1" // Green
	if uptime < 95.0 {
		color = "#e05d44" // Red
	} else if uptime < 99.0 {
		color = "#dfb317" // Yellow
	}

	valueText := fmt.Sprintf("%.2f%%", uptime)
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="108" height="20">
  <linearGradient id="b" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a">
    <rect width="108" height="20" rx="3" fill="#fff"/>
  </mask>
  <g mask="url(#a)">
    <path fill="#555" d="M0 0h52v20H0z"/>
    <path fill="%s" d="M52 0h56v20H52z"/>
    <path fill="url(#b)" d="M0 0h108v20H0z"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="26" y="15" fill="#010101" fill-opacity=".3">uptime</text>
    <text x="26" y="14">uptime</text>
    <text x="79" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="79" y="14">%s</text>
  </g>
</svg>`, color, valueText, valueText)

	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(svg))
}
