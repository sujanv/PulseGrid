package anomaly

import (
	"sync"
	"time"
)

type HourlyBaseline struct {
	HourBuckets [24]*WelfordStats `json:"hour_buckets"`
	mu          sync.RWMutex
}

func NewHourlyBaseline() *HourlyBaseline {
	b := &HourlyBaseline{}
	for i := 0; i < 24; i++ {
		b.HourBuckets[i] = NewWelfordStats()
	}
	return b
}

func (b *HourlyBaseline) Record(t time.Time, latencyMs float64) {
	hour := t.Hour()
	b.HourBuckets[hour].Update(latencyMs)
}

func (b *HourlyBaseline) GetStatsForHour(hour int) (mean float64, stdDev float64) {
	if hour < 0 || hour >= 24 {
		hour = 0
	}
	st := b.HourBuckets[hour]
	return st.Mean, st.StdDev()
}
