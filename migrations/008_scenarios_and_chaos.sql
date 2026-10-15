-- PulseGrid Migration: Scenarios, Chaos Experiments, and Post-Mortems

CREATE TABLE IF NOT EXISTS synthetic_scenarios (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    steps_json JSONB NOT NULL,
    timeout_seconds INT NOT NULL DEFAULT 30,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS scenario_runs (
    id VARCHAR(64) PRIMARY KEY,
    scenario_id VARCHAR(64) REFERENCES synthetic_scenarios(id) ON DELETE CASCADE,
    success BOOLEAN NOT NULL,
    total_duration_ms INT NOT NULL,
    failed_step VARCHAR(255),
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS chaos_experiments (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    target_id VARCHAR(64) NOT NULL,
    fault_type VARCHAR(64) NOT NULL,
    probability DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    status VARCHAR(32) NOT NULL DEFAULT 'SCHEDULED',
    duration_seconds INT NOT NULL DEFAULT 60,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS incident_postmortems (
    incident_id VARCHAR(64) PRIMARY KEY,
    target_id VARCHAR(64) NOT NULL,
    root_cause VARCHAR(128) NOT NULL,
    markdown_content TEXT NOT NULL,
    mttd_seconds INT NOT NULL DEFAULT 0,
    mttr_seconds INT NOT NULL DEFAULT 0,
    sla_burned_pct DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_scenario_runs_scenario_id ON scenario_runs(scenario_id);
CREATE INDEX IF NOT EXISTS idx_chaos_target ON chaos_experiments(target_id);
