package engine

import (
	"sync"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type transitionRecord struct {
	timestamp time.Time
	status    models.CheckStatus
}

type FlapDampener struct {
	mu           sync.RWMutex
	history      map[string][]transitionRecord
	window       time.Duration
	maxSwitches  int
	isFlapping   map[string]bool
}

func NewFlapDampener(window time.Duration, maxSwitches int) *FlapDampener {
	if window <= 0 {
		window = 10 * time.Minute
	}
	if maxSwitches <= 0 {
		maxSwitches = 4
	}
	return &FlapDampener{
		history:     make(map[string][]transitionRecord),
		window:      window,
		maxSwitches: maxSwitches,
		isFlapping:  make(map[string]bool),
	}
}

func (f *FlapDampener) RecordAndCheck(targetID string, status models.CheckStatus) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-f.window)

	records := f.history[targetID]
	var cleaned []transitionRecord
	for _, r := range records {
		if r.timestamp.After(cutoff) {
			cleaned = append(cleaned, r)
		}
	}

	// Only add if status changed
	if len(cleaned) == 0 || cleaned[len(cleaned)-1].status != status {
		cleaned = append(cleaned, transitionRecord{timestamp: now, status: status})
	}
	f.history[targetID] = cleaned

	switchCount := len(cleaned) - 1
	if switchCount < 0 {
		switchCount = 0
	}

	flapping := switchCount >= f.maxSwitches
	f.isFlapping[targetID] = flapping
	return flapping
}

func (f *FlapDampener) IsFlapping(targetID string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.isFlapping[targetID]
}
