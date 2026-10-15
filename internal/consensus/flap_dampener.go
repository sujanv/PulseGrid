package consensus

import (
	"sync"
	"time"
)

type TargetFlapState struct {
	LastStatus       bool      `json:"last_status"`
	FlapScore        float64   `json:"flap_score"`
	IsFlapping       bool      `json:"is_flapping"`
	LastStateChange  time.Time `json:"last_state_change"`
	TransitionCount  int       `json:"transition_count"`
}

type FlapDampener struct {
	mu             sync.Mutex
	states         map[string]*TargetFlapState
	flapThreshold  float64
	clearThreshold float64
	halfLife       time.Duration
}

func NewFlapDampener(flapThreshold, clearThreshold float64, halfLife time.Duration) *FlapDampener {
	if flapThreshold <= 0 {
		flapThreshold = 100.0
	}
	if clearThreshold <= 0 {
		clearThreshold = 40.0
	}
	if halfLife <= 0 {
		halfLife = 60 * time.Second
	}
	return &FlapDampener{
		states:         make(map[string]*TargetFlapState),
		flapThreshold:  flapThreshold,
		clearThreshold: clearThreshold,
		halfLife:       halfLife,
	}
}

func (d *FlapDampener) ProcessStatus(targetID string, isUp bool, now time.Time) (isFlapping bool, suppressAlert bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	state, exists := d.states[targetID]
	if !exists {
		d.states[targetID] = &TargetFlapState{
			LastStatus:      isUp,
			FlapScore:       0,
			IsFlapping:      false,
			LastStateChange: now,
			TransitionCount: 0,
		}
		return false, false
	}

	// Calculate decay
	elapsed := now.Sub(state.LastStateChange)
	decayCycles := float64(elapsed) / float64(d.halfLife)
	decayFactor := 1.0 / (1.0 + decayCycles)
	state.FlapScore *= decayFactor

	if isUp != state.LastStatus {
		// State transition detected: add flap penalty
		state.FlapScore += 35.0
		state.TransitionCount++
		state.LastStatus = isUp
		state.LastStateChange = now
	}

	// Evaluate flapping state
	if !state.IsFlapping && state.FlapScore >= d.flapThreshold {
		state.IsFlapping = true
	} else if state.IsFlapping && state.FlapScore <= d.clearThreshold {
		state.IsFlapping = false
	}

	return state.IsFlapping, state.IsFlapping
}
