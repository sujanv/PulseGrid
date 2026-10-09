CREATE TABLE IF NOT EXISTS incidents (
    id VARCHAR(64) PRIMARY KEY,
    target_id VARCHAR(64) NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
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
