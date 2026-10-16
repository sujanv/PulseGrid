-- PulseGrid Migration: ACME Certificates, On-Call Rotations, Web Vitals, Jitter Profiling

CREATE TABLE IF NOT EXISTS acme_certificates (
    domain VARCHAR(255) PRIMARY KEY,
    status VARCHAR(32) NOT NULL DEFAULT 'ISSUED',
    challenge_type VARCHAR(32) NOT NULL DEFAULT 'http-01',
    fingerprint VARCHAR(64) NOT NULL,
    issued_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    auto_renew BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS oncall_schedules (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    shift_type VARCHAR(32) NOT NULL DEFAULT 'WEEKLY',
    engineers_json JSONB NOT NULL,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS web_vitals_records (
    id SERIAL PRIMARY KEY,
    url VARCHAR(512) NOT NULL,
    lcp_ms INT NOT NULL,
    fid_ms INT NOT NULL,
    cls_score DOUBLE PRECISION NOT NULL,
    ttfb_ms INT NOT NULL,
    overall_pass BOOLEAN NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS jitter_profiles (
    id SERIAL PRIMARY KEY,
    target_host VARCHAR(255) NOT NULL,
    packets_sent INT NOT NULL,
    packets_recv INT NOT NULL,
    loss_pct DOUBLE PRECISION NOT NULL,
    mean_rtt_ms INT NOT NULL,
    jitter_ms INT NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vitals_url ON web_vitals_records(url);
CREATE INDEX IF NOT EXISTS idx_jitter_host ON jitter_profiles(target_host);
