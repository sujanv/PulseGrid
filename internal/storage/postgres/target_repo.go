package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/sujanv/PulseGrid/internal/models"
)

type TargetRepository interface {
	Create(ctx context.Context, target *models.Target) error
	GetByID(ctx context.Context, id string) (*models.Target, error)
	ListAll(ctx context.Context) ([]*models.Target, error)
	ListActive(ctx context.Context) ([]*models.Target, error)
	Update(ctx context.Context, target *models.Target) error
	Delete(ctx context.Context, id string) error
}

type targetRepo struct {
	db *DB
}

func NewTargetRepository(db *DB) TargetRepository {
	return &targetRepo{db: db}
}

func (r *targetRepo) Create(ctx context.Context, t *models.Target) error {
	query := `
		INSERT INTO targets (
			id, name, url, host, port, type, interval_seconds, timeout_seconds,
			method, headers, body, expected_status_codes, expected_keyword,
			failure_threshold, success_threshold, active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`
	headersJSON, _ := json.Marshal(t.Headers)
	intervalSec := int(t.Interval.Seconds())
	timeoutSec := int(t.Timeout.Seconds())

	_, err := r.db.ExecContext(
		ctx, query,
		t.ID, t.Name, t.URL, t.Host, t.Port, string(t.Type),
		intervalSec, timeoutSec, string(t.Method), headersJSON,
		t.Body, pq.Array(t.ExpectedStatusCodes), t.ExpectedKeyword,
		t.FailureThreshold, t.SuccessThreshold, t.Active,
		t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}
	return nil
}

func (r *targetRepo) GetByID(ctx context.Context, id string) (*models.Target, error) {
	query := `
		SELECT id, name, url, host, port, type, interval_seconds, timeout_seconds,
		       method, headers, body, expected_status_codes, expected_keyword,
		       failure_threshold, success_threshold, active, created_at, updated_at
		FROM targets WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return scanTarget(row)
}

func (r *targetRepo) ListAll(ctx context.Context) ([]*models.Target, error) {
	query := `
		SELECT id, name, url, host, port, type, interval_seconds, timeout_seconds,
		       method, headers, body, expected_status_codes, expected_keyword,
		       failure_threshold, success_threshold, active, created_at, updated_at
		FROM targets ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []*models.Target
	for rows.Next() {
		t, err := scanTargetRows(rows)
		if err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func (r *targetRepo) ListActive(ctx context.Context) ([]*models.Target, error) {
	query := `
		SELECT id, name, url, host, port, type, interval_seconds, timeout_seconds,
		       method, headers, body, expected_status_codes, expected_keyword,
		       failure_threshold, success_threshold, active, created_at, updated_at
		FROM targets WHERE active = true ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []*models.Target
	for rows.Next() {
		t, err := scanTargetRows(rows)
		if err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, nil
}

func (r *targetRepo) Update(ctx context.Context, t *models.Target) error {
	query := `
		UPDATE targets SET
			name = $2, url = $3, host = $4, port = $5, type = $6,
			interval_seconds = $7, timeout_seconds = $8, method = $9,
			headers = $10, body = $11, expected_status_codes = $12,
			expected_keyword = $13, failure_threshold = $14,
			success_threshold = $15, active = $16, updated_at = $17
		WHERE id = $1
	`
	headersJSON, _ := json.Marshal(t.Headers)
	intervalSec := int(t.Interval.Seconds())
	timeoutSec := int(t.Timeout.Seconds())
	t.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx, query,
		t.ID, t.Name, t.URL, t.Host, t.Port, string(t.Type),
		intervalSec, timeoutSec, string(t.Method), headersJSON,
		t.Body, pq.Array(t.ExpectedStatusCodes), t.ExpectedKeyword,
		t.FailureThreshold, t.SuccessThreshold, t.Active, t.UpdatedAt,
	)
	return err
}

func (r *targetRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM targets WHERE id = $1", id)
	return err
}

func scanTarget(row *sql.Row) (*models.Target, error) {
	var t models.Target
	var targetType, method, headersRaw string
	var intervalSec, timeoutSec int
	var statusCodes pq.Int64Array

	err := row.Scan(
		&t.ID, &t.Name, &t.URL, &t.Host, &t.Port, &targetType,
		&intervalSec, &timeoutSec, &method, &headersRaw, &t.Body,
		&statusCodes, &t.ExpectedKeyword, &t.FailureThreshold,
		&t.SuccessThreshold, &t.Active, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	t.Type = models.TargetType(targetType)
	t.Method = models.HTTPMethod(method)
	t.Interval = time.Duration(intervalSec) * time.Second
	t.Timeout = time.Duration(timeoutSec) * time.Second
	_ = json.Unmarshal([]byte(headersRaw), &t.Headers)

	t.ExpectedStatusCodes = make([]int, len(statusCodes))
	for i, v := range statusCodes {
		t.ExpectedStatusCodes[i] = int(v)
	}

	return &t, nil
}

func scanTargetRows(rows *sql.Rows) (*models.Target, error) {
	var t models.Target
	var targetType, method, headersRaw string
	var intervalSec, timeoutSec int
	var statusCodes pq.Int64Array

	err := rows.Scan(
		&t.ID, &t.Name, &t.URL, &t.Host, &t.Port, &targetType,
		&intervalSec, &timeoutSec, &method, &headersRaw, &t.Body,
		&statusCodes, &t.ExpectedKeyword, &t.FailureThreshold,
		&t.SuccessThreshold, &t.Active, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	t.Type = models.TargetType(targetType)
	t.Method = models.HTTPMethod(method)
	t.Interval = time.Duration(intervalSec) * time.Second
	t.Timeout = time.Duration(timeoutSec) * time.Second
	_ = json.Unmarshal([]byte(headersRaw), &t.Headers)

	t.ExpectedStatusCodes = make([]int, len(statusCodes))
	for i, v := range statusCodes {
		t.ExpectedStatusCodes[i] = int(v)
	}

	return &t, nil
}
