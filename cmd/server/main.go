package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sujanv/PulseGrid/internal/alert"
	"github.com/sujanv/PulseGrid/internal/api"
	"github.com/sujanv/PulseGrid/internal/checker"
	"github.com/sujanv/PulseGrid/internal/config"
	"github.com/sujanv/PulseGrid/internal/consensus"
	"github.com/sujanv/PulseGrid/internal/engine"
	"github.com/sujanv/PulseGrid/internal/maintenance"
	"github.com/sujanv/PulseGrid/internal/metrics"
	"github.com/sujanv/PulseGrid/internal/models"
	"github.com/sujanv/PulseGrid/internal/pool"
	"github.com/sujanv/PulseGrid/internal/scheduler"
	"github.com/sujanv/PulseGrid/internal/storage/postgres"
	"github.com/sujanv/PulseGrid/internal/storage/redis"
	"github.com/sujanv/PulseGrid/internal/web"
)

func main() {
	log.Println("⚡ Starting PulseGrid Uptime Monitoring Service...")

	cfg := config.Load()

	// 1. Storage layers (PostgreSQL & Redis)
	var db *postgres.DB
	var targetRepo postgres.TargetRepository
	var resultRepo postgres.ResultRepository
	var incidentRepo postgres.IncidentRepository

	pgDB, err := postgres.Connect(cfg.Database)
	if err != nil {
		log.Printf("[Warning] PostgreSQL connection failed: %v (running in volatile memory mode)", err)
	} else {
		db = pgDB
		targetRepo = postgres.NewTargetRepository(db)
		resultRepo = postgres.NewResultRepository(db)
		incidentRepo = postgres.NewIncidentRepository(db)
		log.Println("[Storage] Connected to PostgreSQL and applied migrations.")

		// Start background rollup and pruning worker (every 1 hour)
		rollupWorker := postgres.NewRollupWorker(db, 30)
		rollupWorker.Start(context.Background(), 1*time.Hour)
	}

	var redisClient *redis.Client
	var locker redis.Locker = &redis.NoopLocker{}
	rdb, err := redis.NewClient(cfg.Redis)
	if err != nil {
		log.Printf("[Warning] Redis connection failed: %v (distributed locking disabled)", err)
	} else {
		redisClient = rdb
		locker = redis.NewDistributedLock(redisClient, "pulsegrid-node-1")
		log.Println("[Storage] Connected to Redis.")
	}

	// 2. Health Checker Registry with all protocols
	registry := checker.NewRegistry()

	// 3. Concurrency Worker Pool & Rate Limiter
	workerPool := pool.NewWorkerPool(cfg.Pool.WorkerCount, cfg.Pool.QueueSize)
	rateLimiter := pool.NewPerHostRateLimiter(cfg.Pool.PerHostRateLimit)
	log.Printf("[WorkerPool] Initialized with %d workers, queue capacity %d.", cfg.Pool.WorkerCount, cfg.Pool.QueueSize)

	// 4. Maintenance & Consensus engines
	maintManager := maintenance.NewManager()
	quorumConsensus := consensus.NewQuorumConsensus(60*time.Second, 0.60)

	// 5. Alerting Subsystem
	alertDispatcher := alert.NewDispatcher(5, 500, cfg.Alert.MaxRetries)
	if cfg.Alert.SMTPHost != "" {
		alertDispatcher.RegisterChannel(alert.NewEmailChannel(
			cfg.Alert.SMTPHost, cfg.Alert.SMTPPort, cfg.Alert.SMTPUser,
			cfg.Alert.SMTPPass, cfg.Alert.FromEmail, "admin@pulsegrid.io",
		))
	}

	// 6. Incident Detection Engine
	detector := engine.NewIncidentDetector(incidentRepo, func(event *models.AlertEvent) {
		log.Printf("[Alert] %s: %s", event.EventType, event.Message)
		alertDispatcher.Dispatch(event)
	})
	detector.SetMaintenanceManager(maintManager)
	detector.SetConsensusEngine(quorumConsensus)

	// 7. Scheduler Engine with Jitter
	sched := scheduler.NewScheduler(
		workerPool,
		registry,
		rateLimiter,
		cfg.Scheduler.JitterPercent,
		func(target *models.Target, res *models.CheckResult, err error) {
			if res != nil {
				metrics.RecordResult(target, res)
				if resultRepo != nil {
					_ = resultRepo.SaveResult(context.Background(), res)
				}
				detector.ProcessResult(context.Background(), target, res)
			}
		},
	)

	// Load targets if repo available
	if targetRepo != nil {
		targets, err := targetRepo.ListActive(context.Background())
		if err == nil {
			for _, t := range targets {
				sched.RegisterTarget(t)
			}
			log.Printf("[Scheduler] Loaded %d active targets from database.", len(targets))
		}
	}

	// 8. HTTP API Server & UI
	router := api.NewRouter()
	router.HandleFunc("/", web.Handler())
	router.HandleFunc("/dashboard", web.Handler())
	router.Handle("/metrics", promhttp.Handler())

	statusPage := web.NewStatusPageHandler(targetRepo, resultRepo, incidentRepo)
	router.HandleFunc("/status", statusPage.ServeHTTP)

	badgeHandler := web.NewBadgeHandler(resultRepo)
	router.HandleFunc("/badge/", badgeHandler.ServeBadge)

	healthHandler := api.NewHealthHandler(db, redisClient, workerPool)
	router.HandleFunc("/healthz", healthHandler.Healthz)
	router.HandleFunc("/readyz", healthHandler.Readyz)

	maintHandler := api.NewMaintenanceHandler(maintManager)
	router.HandleFunc("/api/v1/maintenance", maintHandler.HandleWindows)
	router.HandleFunc("/api/v1/maintenance/", maintHandler.HandleWindows)

	if targetRepo != nil {
		targetHandler := api.NewTargetHandler(targetRepo)
		router.HandleFunc("/api/v1/targets", targetHandler.HandleTargets)
		router.HandleFunc("/api/v1/targets/", targetHandler.HandleTargetByID)
	}

	if resultRepo != nil {
		resultHandler := api.NewResultHandler(resultRepo)
		router.HandleFunc("/api/v1/targets/", func(w http.ResponseWriter, r *http.Request) {
			if len(r.URL.Path) > 8 && r.URL.Path[len(r.URL.Path)-8:] == "/results" {
				resultHandler.HandleResults(w, r)
			} else if len(r.URL.Path) > 6 && r.URL.Path[len(r.URL.Path)-6:] == "/stats" {
				resultHandler.HandleStats(w, r)
			}
		})

		reportHandler := api.NewReportHandler(targetRepo, resultRepo)
		router.HandleFunc("/api/v1/reports/sla", reportHandler.HandleSLAReport)
		router.HandleFunc("/api/v1/reports/export.csv", reportHandler.HandleCSVExport)
	}

	if incidentRepo != nil {
		incidentHandler := api.NewIncidentHandler(incidentRepo)
		router.HandleFunc("/api/v1/incidents", incidentHandler.HandleIncidents)
		router.HandleFunc("/api/v1/incidents/active", incidentHandler.HandleActiveIncidents)
	}

	var handler http.Handler = router
	handler = api.CORSMiddleware(handler)
	handler = api.RecoveryMiddleware(handler)
	handler = api.LoggingMiddleware(handler)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("[Server] PulseGrid HTTP API & UI listening at http://%s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 9. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Shutdown] Initiating graceful shutdown...")
	sched.Stop()
	_ = workerPool.Stop(5 * time.Second)
	alertDispatcher.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[Shutdown] Server shutdown error: %v", err)
	}

	if db != nil {
		_ = db.Close()
	}
	if redisClient != nil {
		_ = redisClient.Close()
	}
	log.Println("[Shutdown] PulseGrid cleanly stopped.")
}
