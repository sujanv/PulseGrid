-- PulseGrid Migration: HAR Archives, mTLS Profiles, Alert Groups, Canary Evaluations

CREATE TABLE IF NOT EXISTS har_archives (
    id VARCHAR(64) PRIMARY KEY,
    scenario_id VARCHAR(64) NOT NULL,
    archive_json JSONB NOT NULL,
    total_entries INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS mtls_profiles (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    client_cert_pem TEXT NOT NULL,
    client_key_pem TEXT NOT NULL,
    root_ca_pem TEXT,
    insecure_skip_verify BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS alert_groups (
    fingerprint VARCHAR(64) PRIMARY KEY,
    provider VARCHAR(64) NOT NULL,
    region VARCHAR(64) NOT NULL,
    zone VARCHAR(64) NOT NULL,
    service VARCHAR(64) NOT NULL,
    target_count INT NOT NULL DEFAULT 1,
    first_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS canary_evaluations (
    id VARCHAR(64) PRIMARY KEY,
    baseline_id VARCHAR(64) NOT NULL,
    canary_id VARCHAR(64) NOT NULL,
    decision VARCHAR(32) NOT NULL,
    confidence_pct DOUBLE PRECISION NOT NULL,
    reason TEXT NOT NULL,
    evaluated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_har_scenario ON har_archives(scenario_id);
CREATE INDEX IF NOT EXISTS idx_canary_eval ON canary_evaluations(baseline_id, canary_id);
