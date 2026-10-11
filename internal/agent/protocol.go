package agent

import (
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type CheckJob struct {
	JobID    string         `json:"job_id"`
	Target   *models.Target `json:"target"`
	IssuedAt time.Time      `json:"issued_at"`
}

type JobResult struct {
	JobID       string              `json:"job_id"`
	AgentID     string              `json:"agent_id"`
	Region      string              `json:"region"`
	CheckResult *models.CheckResult `json:"check_result"`
}
