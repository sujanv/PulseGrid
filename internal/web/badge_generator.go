package web

import (
	"fmt"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type BadgeGenerator struct{}

func NewBadgeGenerator() *BadgeGenerator {
	return &BadgeGenerator{}
}

func (b *BadgeGenerator) GenerateSVG(status models.CheckStatus, latency time.Duration, uptimePct float64) string {
	color := "#4c1" // Green (Up)
	statusText := "up"

	if status == models.StatusDown {
		color = "#e05d44" // Red (Down)
		statusText = "down"
	} else if status == models.StatusDegraded {
		color = "#dfb317" // Yellow (Degraded)
		statusText = "degraded"
	}

	latencyMs := latency.Milliseconds()
	valueText := fmt.Sprintf("%s | %.1f%% (%dms)", statusText, uptimePct, latencyMs)

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="180" height="20">
  <linearGradient id="b" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a">
    <rect width="180" height="20" rx="3" fill="#fff"/>
  </mask>
  <g mask="url(#a)">
    <path fill="#555" d="M0 0h60v20H0z"/>
    <path fill="%s" d="M60 0h120v20H60z"/>
    <path fill="url(#b)" d="M0 0h180v20H0z"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="30" y="15" fill="#010101" fill-opacity=".3">uptime</text>
    <text x="30" y="14">uptime</text>
    <text x="120" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="120" y="14">%s</text>
  </g>
</svg>`, color, valueText, valueText)
}
