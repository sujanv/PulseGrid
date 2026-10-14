package postgres

import (
	"database/sql"
	"fmt"
	"log"
)

var migrationScripts = []struct {
	name string
	sql  string
}{
	{
		name: "001_create_targets",
		sql: `
CREATE TABLE IF NOT EXISTS targets (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    host VARCHAR(255) DEFAULT '',
    port INT DEFAULT 0,
    type VARCHAR(32) NOT NULL DEFAULT 'HTTP',
    interval_seconds INT NOT NULL DEFAULT 30,
    timeout_seconds INT NOT NULL DEFAULT 10,
    method VARCHAR(16) DEFAULT 'GET',
    headers JSONB DEFAULT '{}'::jsonb,
    body TEXT DEFAULT '',
    expected_status_codes INT[] DEFAULT ARRAY[200, 201, 204],
    expected_keyword VARCHAR(255) DEFAULT '',
    failure_threshold INT NOT NULL DEFAULT 3,
    success_threshold INT NOT NULL DEFAULT 2,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_targets_active ON targets (active);
CREATE INDEX IF NOT EXISTS idx_targets_type ON targets (type);
`,
	},
	{
		name: "002_create_results",
		sql: `
CREATE TABLE IF NOT EXISTS check_results (
    id VARCHAR(64) PRIMARY KEY,
    target_id VARCHAR(64) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(32) NOT NULL,
    status_code INT DEFAULT 0,
    response_bytes BIGINT DEFAULT 0,
    latency_ms DOUBLE PRECISION NOT NULL,
    dns_lookup_ms DOUBLE PRECISION DEFAULT 0,
    tcp_connect_ms DOUBLE PRECISION DEFAULT 0,
    tls_handshake_ms DOUBLE PRECISION DEFAULT 0,
    ttfb_ms DOUBLE PRECISION DEFAULT 0,
    content_transfer_ms DOUBLE PRECISION DEFAULT 0,
    error_message TEXT DEFAULT '',
    cert_days_left INT DEFAULT 0,
    resolved_ips TEXT[] DEFAULT ARRAY[]::text[]
);
CREATE INDEX IF NOT EXISTS idx_results_target_time ON check_results (target_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_results_timestamp ON check_results (timestamp DESC);
`,
	},
	{
		name: "003_create_incidents",
		sql: `
CREATE TABLE IF NOT EXISTS incidents (
    id VARCHAR(64) PRIMARY KEY,
    target_id VARCHAR(64) NOT NULL,
    target_name VARCHAR(255) NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INT DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    root_cause TEXT DEFAULT '',
    consecutive_failures INT DEFAULT 1,
    severity VARCHAR(32) DEFAULT 'CRITICAL'
);
CREATE INDEX IF NOT EXISTS idx_incidents_target_status ON incidents (target_id, status);
CREATE INDEX IF NOT EXISTS idx_incidents_started_at ON incidents (started_at DESC);
`,
	},
	{
		name: "004_create_alerts",
		sql: `
CREATE TABLE IF NOT EXISTS alert_channels (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(32) NOT NULL,
    destination TEXT NOT NULL,
    secret_key TEXT DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS alert_logs (
    id VARCHAR(64) PRIMARY KEY,
    event_id VARCHAR(64) NOT NULL,
    channel_id VARCHAR(64) NOT NULL,
    channel_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    error_message TEXT DEFAULT '',
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_alert_logs_event ON alert_logs (event_id);
`,
	},
	{
		name: "005_create_maintenance_and_rollups",
		sql: `
CREATE TABLE IF NOT EXISTS maintenance_windows (
    id VARCHAR(64) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    target_ids TEXT[] DEFAULT ARRAY[]::text[],
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS daily_rollups (
    target_id VARCHAR(64) NOT NULL,
    rollup_date DATE NOT NULL,
    total_checks BIGINT NOT NULL DEFAULT 0,
    up_checks BIGINT NOT NULL DEFAULT 0,
    avg_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    p95_latency_ms DOUBLE PRECISION NOT NULL DEFAULT 0,
    PRIMARY KEY (target_id, rollup_date)
);
CREATE INDEX IF NOT EXISTS idx_rollups_date ON daily_rollups (rollup_date DESC);
`,
	},
	{
		name: "007_create_remediation_actions",
		sql: `
CREATE TABLE IF NOT EXISTS remediation_rules (
    id VARCHAR(64) PRIMARY KEY,
    target_id VARCHAR(64) NOT NULL,
    action_type VARCHAR(32) NOT NULL DEFAULT 'WEBHOOK',
    action_url TEXT NOT NULL,
    cooldown_seconds INT NOT NULL DEFAULT 900,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS remediation_audits (
    id VARCHAR(64) PRIMARY KEY,
    target_id VARCHAR(64) NOT NULL,
    action_url TEXT NOT NULL,
    status_code INT NOT NULL,
    duration_ms DOUBLE PRECISION NOT NULL,
    success BOOLEAN NOT NULL,
    error_message TEXT DEFAULT '',
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`,
	},
}

func RunMigrations(db *sql.DB) error {
	for _, m := range migrationScripts {
		if _, err := db.Exec(m.sql); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.name, err)
		}
		log.Printf("[Storage] Migration applied: %s", m.name)
	}
	return nil
}
