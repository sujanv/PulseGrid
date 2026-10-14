package remediation

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type ExecutionAuditLog struct {
	ID         string        `json:"id"`
	TargetID   string        `json:"target_id"`
	ActionURL  string        `json:"action_url"`
	StatusCode int           `json:"status_code"`
	Duration   time.Duration `json:"duration"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
	ExecutedAt time.Time     `json:"executed_at"`
}

type AuditLogger struct {
	mu   sync.RWMutex
	logs map[string][]ExecutionAuditLog
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logs: make(map[string][]ExecutionAuditLog),
	}
}

func (l *AuditLogger) Log(targetID, url string, code int, dur time.Duration, success bool, errStr string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := ExecutionAuditLog{
		ID:         uuid.New().String(),
		TargetID:   targetID,
		ActionURL:  url,
		StatusCode: code,
		Duration:   dur,
		Success:    success,
		Error:      errStr,
		ExecutedAt: time.Now().UTC(),
	}
	l.logs[targetID] = append(l.logs[targetID], entry)
	if len(l.logs[targetID]) > 50 {
		l.logs[targetID] = l.logs[targetID][1:]
	}
}

func (l *AuditLogger) GetLogs(targetID string) []ExecutionAuditLog {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.logs[targetID]
}
