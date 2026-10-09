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
    channel_id VARCHAR(64) NOT NULL REFERENCES alert_channels(id) ON DELETE CASCADE,
    channel_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    error_message TEXT DEFAULT '',
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alert_logs_event ON alert_logs (event_id);
CREATE INDEX IF NOT EXISTS idx_alert_logs_sent ON alert_logs (sent_at DESC);
