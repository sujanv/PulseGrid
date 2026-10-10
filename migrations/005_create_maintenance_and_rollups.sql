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
