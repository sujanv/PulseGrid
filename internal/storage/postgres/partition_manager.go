package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PartitionManager struct {
	db *sql.DB
}

func NewPartitionManager(db *sql.DB) *PartitionManager {
	return &PartitionManager{db: db}
}

func (m *PartitionManager) EnsurePartitionsForNextMonths(ctx context.Context, monthsAhead int) error {
	if m.db == nil {
		return nil
	}

	now := time.Now().UTC()
	for i := 0; i < monthsAhead; i++ {
		targetMonth := now.AddDate(0, i, 0)
		year := targetMonth.Year()
		month := int(targetMonth.Month())

		query := fmt.Sprintf("SELECT create_check_results_partition(%d, %d);", year, month)
		if _, err := m.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed creating partition for %d-%02d: %w", year, month, err)
		}
	}
	return nil
}

func (m *PartitionManager) DropOldPartitions(ctx context.Context, retentionDays int) error {
	if m.db == nil {
		return nil
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
	year := cutoff.Year()
	month := int(cutoff.Month())
	oldTableName := fmt.Sprintf("check_results_%d_%02d", year, month)

	dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s;", oldTableName)
	_, err := m.db.ExecContext(ctx, dropQuery)
	return err
}
