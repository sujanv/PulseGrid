package api

import (
	"context"
	"net/http"
	"time"

	"github.com/sujanv/PulseGrid/internal/pool"
	"github.com/sujanv/PulseGrid/internal/storage/postgres"
	"github.com/sujanv/PulseGrid/internal/storage/redis"
)

type HealthHandler struct {
	db     *postgres.DB
	redis  *redis.Client
	pool   *pool.WorkerPool
}

func NewHealthHandler(db *postgres.DB, r *redis.Client, p *pool.WorkerPool) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: r,
		pool:  p,
	}
}

func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	JSONResponse(w, http.StatusOK, map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allHealthy := true

	if h.db != nil {
		if err := h.db.HealthCheck(ctx); err != nil {
			checks["postgres"] = "DOWN: " + err.Error()
			allHealthy = false
		} else {
			checks["postgres"] = "UP"
		}
	} else {
		checks["postgres"] = "NOT_CONFIGURED"
	}

	if h.redis != nil {
		if err := h.redis.Ping(ctx); err != nil {
			checks["redis"] = "DOWN: " + err.Error()
			allHealthy = false
		} else {
			checks["redis"] = "UP"
		}
	} else {
		checks["redis"] = "NOT_CONFIGURED"
	}

	if h.pool != nil {
		stats := h.pool.Stats()
		checks["worker_pool"] = "UP"
		checks["active_workers"] = string(rune(stats.ActiveWorkers))
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	JSONResponse(w, status, map[string]interface{}{
		"status": map[bool]string{true: "ready", false: "not_ready"}[allHealthy],
		"checks": checks,
	})
}
