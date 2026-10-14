package postgres

import (
	"context"
	"log"
	"time"
)

type RollupWorker struct {
	db            *DB
	retentionDays int
}

func NewRollupWorker(db *DB, retentionDays int) *RollupWorker {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	return &RollupWorker{
		db:            db,
		retentionDays: retentionDays,
	}
}

func (w *RollupWorker) Start(ctx context.Context, interval time.Duration) {
	if w.db == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.ExecuteRollup(ctx); err != nil {
					log.Printf("[RollupWorker] Rollup aggregation failed: %v", err)
				}
				if err := w.PruneOldData(ctx); err != nil {
					log.Printf("[RollupWorker] Prune failed: %v", err)
				}
			}
		}
	}()
}

func (w *RollupWorker) ExecuteRollup(ctx context.Context) error {
	query := `
		INSERT INTO daily_rollups (target_id, rollup_date, total_checks, up_checks, avg_latency_ms, p95_latency_ms)
		SELECT
			target_id,
			DATE(timestamp) as rollup_date,
			COUNT(*) as total_checks,
			COUNT(CASE WHEN status = 'UP' THEN 1 END) as up_checks,
			COALESCE(AVG(latency_ms), 0) as avg_latency_ms,
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms), 0) as p95_latency_ms
		FROM check_results
		WHERE timestamp >= NOW() - INTERVAL '2 days'
		GROUP BY target_id, DATE(timestamp)
		ON CONFLICT (target_id, rollup_date) DO UPDATE SET
			total_checks = EXCLUDED.total_checks,
			up_checks = EXCLUDED.up_checks,
			avg_latency_ms = EXCLUDED.avg_latency_ms,
			p95_latency_ms = EXCLUDED.p95_latency_ms;
	`
	_, err := w.db.ExecContext(ctx, query)
	return err
}

func (w *RollupWorker) PruneOldData(ctx context.Context) error {
	cutoff := time.Now().AddDate(0, 0, -w.retentionDays)
	query := `DELETE FROM check_results WHERE timestamp < $1`
	res, err := w.db.ExecContext(ctx, query, cutoff)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows > 0 {
		log.Printf("[RollupWorker] Pruned %d raw check records older than %d days", rows, w.retentionDays)
	}
	return nil
}
