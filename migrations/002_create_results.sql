CREATE TABLE IF NOT EXISTS check_results (
    id VARCHAR(64) PRIMARY KEY,
    target_id VARCHAR(64) NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
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
CREATE INDEX IF NOT EXISTS idx_results_status ON check_results (status);
