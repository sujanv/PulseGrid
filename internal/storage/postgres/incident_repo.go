package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type IncidentRepository interface {
	CreateIncident(ctx context.Context, inc *models.Incident) error
	ResolveIncident(ctx context.Context, id string, resolvedAt time.Time, duration time.Duration) error
	GetActiveByTarget(ctx context.Context, targetID string) (*models.Incident, error)
	ListIncidents(ctx context.Context, limit int, offset int) ([]*models.Incident, error)
	ListActiveIncidents(ctx context.Context) ([]*models.Incident, error)
}

type incidentRepo struct {
	db *DB
}

func NewIncidentRepository(db *DB) IncidentRepository {
	return &incidentRepo{db: db}
}

func (r *incidentRepo) CreateIncident(ctx context.Context, inc *models.Incident) error {
	query := `
		INSERT INTO incidents (
			id, target_id, target_name, started_at, status, root_cause,
			consecutive_failures, severity
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		inc.ID, inc.TargetID, inc.TargetName, inc.StartedAt,
		string(inc.Status), inc.RootCause, inc.ConsecutiveFailures, string(inc.Severity),
	)
	return err
}

func (r *incidentRepo) ResolveIncident(ctx context.Context, id string, resolvedAt time.Time, duration time.Duration) error {
	durationSec := int(duration.Seconds())
	query := `
		UPDATE incidents
		SET status = 'RESOLVED', resolved_at = $2, duration_seconds = $3
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, resolvedAt, durationSec)
	return err
}

func (r *incidentRepo) GetActiveByTarget(ctx context.Context, targetID string) (*models.Incident, error) {
	query := `
		SELECT id, target_id, target_name, started_at, resolved_at, duration_seconds,
		       status, root_cause, consecutive_failures, severity
		FROM incidents
		WHERE target_id = $1 AND status != 'RESOLVED'
		ORDER BY started_at DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, targetID)
	return scanIncident(row)
}

func (r *incidentRepo) ListIncidents(ctx context.Context, limit int, offset int) ([]*models.Incident, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, target_id, target_name, started_at, resolved_at, duration_seconds,
		       status, root_cause, consecutive_failures, severity
		FROM incidents
		ORDER BY started_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*models.Incident
	for rows.Next() {
		inc, err := scanIncidentRows(rows)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, nil
}

func (r *incidentRepo) ListActiveIncidents(ctx context.Context) ([]*models.Incident, error) {
	query := `
		SELECT id, target_id, target_name, started_at, resolved_at, duration_seconds,
		       status, root_cause, consecutive_failures, severity
		FROM incidents
		WHERE status != 'RESOLVED'
		ORDER BY started_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*models.Incident
	for rows.Next() {
		inc, err := scanIncidentRows(rows)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, nil
}

func scanIncident(row *sql.Row) (*models.Incident, error) {
	var inc models.Incident
	var statusStr, severityStr string
	var durationSec int
	var resolvedAt sql.NullTime

	err := row.Scan(
		&inc.ID, &inc.TargetID, &inc.TargetName, &inc.StartedAt,
		&resolvedAt, &durationSec, &statusStr, &inc.RootCause,
		&inc.ConsecutiveFailures, &severityStr,
	)
	if err != nil {
		return nil, err
	}

	inc.Status = models.IncidentStatus(statusStr)
	inc.Severity = models.Severity(severityStr)
	inc.Duration = time.Duration(durationSec) * time.Second
	if resolvedAt.Valid {
		inc.ResolvedAt = &resolvedAt.Time
	}
	return &inc, nil
}

func scanIncidentRows(rows *sql.Rows) (*models.Incident, error) {
	var inc models.Incident
	var statusStr, severityStr string
	var durationSec int
	var resolvedAt sql.NullTime

	err := rows.Scan(
		&inc.ID, &inc.TargetID, &inc.TargetName, &inc.StartedAt,
		&resolvedAt, &durationSec, &statusStr, &inc.RootCause,
		&inc.ConsecutiveFailures, &severityStr,
	)
	if err != nil {
		return nil, err
	}

	inc.Status = models.IncidentStatus(statusStr)
	inc.Severity = models.Severity(severityStr)
	inc.Duration = time.Duration(durationSec) * time.Second
	if resolvedAt.Valid {
		inc.ResolvedAt = &resolvedAt.Time
	}
	return &inc, nil
}
