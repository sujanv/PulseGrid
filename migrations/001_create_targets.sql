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
