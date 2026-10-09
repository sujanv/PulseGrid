package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/sujanv/PulseGrid/internal/models"
)

type LatencyStats struct {
	AvgMs float64 `json:"avg_ms"`
	MinMs float64 `json:"min_ms"`
	MaxMs float64 `json:"max_ms"`
	P50Ms float64 `json:"p50_ms"`
	P95Ms float64 `json:"p95_ms"`
	P99Ms float64 `json:"p99_ms"`
}

type ResultRepository interface {
	SaveResult(ctx context.Context, res *models.CheckResult) error
	GetRecentByTarget(ctx context.Context, targetID string, limit int) ([]*models.CheckResult, error)
	GetLatencyStats(ctx context.Context, targetID string, window time.Duration) (*LatencyStats, error)
	CalculateUptime(ctx context.Context, targetID string, window time.Duration) (float64, error)
}

type resultRepo struct {
	db *DB
}

func NewResultRepository(db *DB) ResultRepository {
	return &resultRepo{db: db}
}

func (r *resultRepo) SaveResult(ctx context.Context, res *models.CheckResult) error {
	query := `
		INSERT INTO check_results (
			id, target_id, timestamp, status, status_code, response_bytes,
			latency_ms, dns_lookup_ms, tcp_connect_ms, tls_handshake_ms,
			ttfb_ms, content_transfer_ms, error_message, cert_days_left, resolved_ips
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	latencyMs := float64(res.Latency.Microseconds()) / 1000.0
	dnsMs := float64(res.Breakdown.DNSLookup.Microseconds()) / 1000.0
	tcpMs := float64(res.Breakdown.TCPConnect.Microseconds()) / 1000.0
	tlsMs := float64(res.Breakdown.TLSHandshake.Microseconds()) / 1000.0
	ttfbMs := float64(res.Breakdown.TTFB.Microseconds()) / 1000.0
	transferMs := float64(res.Breakdown.ContentTransfer.Microseconds()) / 1000.0

	_, err := r.db.ExecContext(
		ctx, query,
		res.ID, res.TargetID, res.Timestamp, string(res.Status),
		res.StatusCode, res.ResponseBytes, latencyMs, dnsMs, tcpMs, tlsMs,
		ttfbMs, transferMs, res.ErrorMessage, res.CertificateDaysLeft, pq.Array(res.ResolvedIPs),
	)
	if err != nil {
		return fmt.Errorf("failed to save check result: %w", err)
	}
	return nil
}

func (r *resultRepo) GetRecentByTarget(ctx context.Context, targetID string, limit int) ([]*models.CheckResult, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, target_id, timestamp, status, status_code, response_bytes,
		       latency_ms, dns_lookup_ms, tcp_connect_ms, tls_handshake_ms,
		       ttfb_ms, content_transfer_ms, error_message, cert_days_left, resolved_ips
		FROM check_results
		WHERE target_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, targetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.CheckResult
	for rows.Next() {
		var res models.CheckResult
		var statusStr string
		var latencyMs, dnsMs, tcpMs, tlsMs, ttfbMs, transferMs float64
		var ips pq.StringArray

		err := rows.Scan(
			&res.ID, &res.TargetID, &res.Timestamp, &statusStr, &res.StatusCode,
			&res.ResponseBytes, &latencyMs, &dnsMs, &tcpMs, &tlsMs,
			&ttfbMs, &transferMs, &res.ErrorMessage, &res.CertificateDaysLeft, &ips,
		)
		if err != nil {
			return nil, err
		}

		res.Status = models.CheckStatus(statusStr)
		res.Latency = time.Duration(latencyMs * float64(time.Millisecond))
		res.Breakdown = models.LatencyBreakdown{
			DNSLookup:       time.Duration(dnsMs * float64(time.Millisecond)),
			TCPConnect:      time.Duration(tcpMs * float64(time.Millisecond)),
			TLSHandshake:    time.Duration(tlsMs * float64(time.Millisecond)),
			TTFB:            time.Duration(ttfbMs * float64(time.Millisecond)),
			ContentTransfer: time.Duration(transferMs * float64(time.Millisecond)),
			Total:           res.Latency,
		}
		res.ResolvedIPs = []string(ips)
		results = append(results, &res)
	}
	return results, nil
}

func (r *resultRepo) GetLatencyStats(ctx context.Context, targetID string, window time.Duration) (*LatencyStats, error) {
	since := time.Now().Add(-window)
	query := `
		SELECT
			COALESCE(AVG(latency_ms), 0),
			COALESCE(MIN(latency_ms), 0),
			COALESCE(MAX(latency_ms), 0),
			COALESCE(PERCENTILE_CONT(0.50) WITHIN GROUP (ORDER BY latency_ms), 0),
			COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms), 0),
			COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY latency_ms), 0)
		FROM check_results
		WHERE target_id = $1 AND timestamp >= $2
	`
	var stats LatencyStats
	err := r.db.QueryRowContext(ctx, query, targetID, since).Scan(
		&stats.AvgMs, &stats.MinMs, &stats.MaxMs,
		&stats.P50Ms, &stats.P95Ms, &stats.P99Ms,
	)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *resultRepo) CalculateUptime(ctx context.Context, targetID string, window time.Duration) (float64, error) {
	since := time.Now().Add(-window)
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'UP' THEN 1 END) as up_count
		FROM check_results
		WHERE target_id = $1 AND timestamp >= $2
	`
	var total, upCount int64
	err := r.db.QueryRowContext(ctx, query, targetID, since).Scan(&total, &upCount)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 100.0, nil
	}
	return (float64(upCount) / float64(total)) * 100.0, nil
}
