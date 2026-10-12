package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/sujanv/PulseGrid/internal/config"
)

type DB struct {
	*sql.DB
	stmtCache map[string]*sql.Stmt
	stmtMu    sync.RWMutex
}

func Connect(cfg config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database url: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{
		DB:        db,
		stmtCache: make(map[string]*sql.Stmt),
	}, nil
}

func (db *DB) HealthCheck(ctx context.Context) error {
	return db.PingContext(ctx)
}

func (db *DB) PrepareCached(query string) (*sql.Stmt, error) {
	db.stmtMu.RLock()
	stmt, exists := db.stmtCache[query]
	db.stmtMu.RUnlock()
	if exists {
		return stmt, nil
	}

	db.stmtMu.Lock()
	defer db.stmtMu.Unlock()
	if stmt, exists = db.stmtCache[query]; exists {
		return stmt, nil
	}

	newStmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	db.stmtCache[query] = newStmt
	return newStmt, nil
}
