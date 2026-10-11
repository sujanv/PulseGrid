package anomaly

import (
	"math"
	"sync"
)

type WelfordStats struct {
	Count  int64   `json:"count"`
	Mean   float64 `json:"mean"`
	M2     float64 `json:"m2"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	mu     sync.RWMutex
}

func NewWelfordStats() *WelfordStats {
	return &WelfordStats{
		Min: math.MaxFloat64,
		Max: -math.MaxFloat64,
	}
}

func (w *WelfordStats) Update(val float64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.Count++
	delta := val - w.Mean
	w.Mean += delta / float64(w.Count)
	delta2 := val - w.Mean
	w.M2 += delta * delta2

	if val < w.Min {
		w.Min = val
	}
	if val > w.Max {
		w.Max = val
	}
}

func (w *WelfordStats) Variance() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.Count < 2 {
		return 0.0
	}
	return w.M2 / float64(w.Count-1)
}

func (w *WelfordStats) StdDev() float64 {
	return math.Sqrt(w.Variance())
}

func (w *WelfordStats) ZScore(val float64) float64 {
	std := w.StdDev()
	if std <= 0.0001 {
		return 0.0
	}
	w.mu.RLock()
	defer w.mu.RUnlock()
	return (val - w.Mean) / std
}
