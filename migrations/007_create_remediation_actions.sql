CREATE TABLE IF NOT EXISTS remediation_rules (
    id VARCHAR(64) PRIMARY KEY,
    target_id VARCHAR(64) NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
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

CREATE INDEX IF NOT EXISTS idx_remediation_target ON remediation_audits (target_id, executed_at DESC);
