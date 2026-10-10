package api

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/sujanv/PulseGrid/internal/storage/postgres"
)

type ReportHandler struct {
	targetRepo postgres.TargetRepository
	resultRepo postgres.ResultRepository
}

func NewReportHandler(t postgres.TargetRepository, r postgres.ResultRepository) *ReportHandler {
	return &ReportHandler{
		targetRepo: t,
		resultRepo: r,
	}
}

func (h *ReportHandler) HandleSLAReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	targets, err := h.targetRepo.ListAll(ctx)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	window := 30 * 24 * time.Hour
	type Item struct {
		TargetID   string  `json:"target_id"`
		TargetName string  `json:"target_name"`
		UptimePct  float64 `json:"uptime_pct"`
		AvgLatency float64 `json:"avg_latency_ms"`
		P95Latency float64 `json:"p95_latency_ms"`
	}

	var report []Item
	for _, t := range targets {
		uptime, _ := h.resultRepo.CalculateUptime(ctx, t.ID, window)
		stats, _ := h.resultRepo.GetLatencyStats(ctx, t.ID, window)
		avgLat, p95Lat := 0.0, 0.0
		if stats != nil {
			avgLat = stats.AvgMs
			p95Lat = stats.P95Ms
		}
		report = append(report, Item{
			TargetID:   t.ID,
			TargetName: t.Name,
			UptimePct:  uptime,
			AvgLatency: avgLat,
			P95Latency: p95Lat,
		})
	}

	JSONResponse(w, http.StatusOK, map[string]interface{}{
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"window_days":  30,
		"targets":      report,
	})
}

func (h *ReportHandler) HandleCSVExport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	targets, err := h.targetRepo.ListAll(ctx)
	if err != nil {
		ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	window := 30 * 24 * time.Hour
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	_ = writer.Write([]string{"Target ID", "Name", "Type", "Endpoint", "Uptime % (30d)", "Avg Latency (ms)", "P95 Latency (ms)"})

	for _, t := range targets {
		uptime, _ := h.resultRepo.CalculateUptime(ctx, t.ID, window)
		stats, _ := h.resultRepo.GetLatencyStats(ctx, t.ID, window)
		avgLat, p95Lat := "0.00", "0.00"
		if stats != nil {
			avgLat = fmt.Sprintf("%.2f", stats.AvgMs)
			p95Lat = fmt.Sprintf("%.2f", stats.P95Ms)
		}
		_ = writer.Write([]string{
			t.ID, t.Name, string(t.Type), t.URL,
			fmt.Sprintf("%.2f", uptime), avgLat, p95Lat,
		})
	}
	writer.Flush()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=pulsegrid_sla_report.csv")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
