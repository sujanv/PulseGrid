package consensus

import (
	"sync"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type RegionalCheckResult struct {
	Region    string
	Status    models.CheckStatus
	Timestamp time.Time
}

type QuorumConsensus struct {
	mu           sync.RWMutex
	results      map[string][]RegionalCheckResult
	window       time.Duration
	quorumRatio  float64 // e.g. 0.67 (2/3 majority required)
}

func NewQuorumConsensus(window time.Duration, quorumRatio float64) *QuorumConsensus {
	if window <= 0 {
		window = 90 * time.Second
	}
	if quorumRatio <= 0 {
		quorumRatio = 0.50
	}
	return &QuorumConsensus{
		results:     make(map[string][]RegionalCheckResult),
		window:      window,
		quorumRatio: quorumRatio,
	}
}

func (q *QuorumConsensus) Record(targetID, region string, status models.CheckStatus) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-q.window)

	list := q.results[targetID]
	var cleaned []RegionalCheckResult
	for _, r := range list {
		if r.Timestamp.After(cutoff) && r.Region != region {
			cleaned = append(cleaned, r)
		}
	}
	cleaned = append(cleaned, RegionalCheckResult{
		Region:    region,
		Status:    status,
		Timestamp: now,
	})
	q.results[targetID] = cleaned
}

func (q *QuorumConsensus) IsOutageConfirmed(targetID string) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	records := q.results[targetID]
	if len(records) == 0 {
		return false
	}

	downCount := 0
	for _, r := range records {
		if r.Status == models.StatusDown {
			downCount++
		}
	}

	ratio := float64(downCount) / float64(len(records))
	return ratio >= q.quorumRatio
}
