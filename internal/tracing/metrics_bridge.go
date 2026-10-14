package tracing

import (
	"sync"
	"time"
)

type MetricRecord struct {
	MetricName string            `json:"metric_name"`
	Value      float64           `json:"value"`
	Labels     map[string]string `json:"labels"`
	Timestamp  time.Time         `json:"timestamp"`
}

type MetricsBridge struct {
	mu      sync.RWMutex
	records []MetricRecord
}

func NewMetricsBridge() *MetricsBridge {
	return &MetricsBridge{}
}

func (b *MetricsBridge) Record(name string, val float64, labels map[string]string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.records = append(b.records, MetricRecord{
		MetricName: name,
		Value:      val,
		Labels:     labels,
		Timestamp:  time.Now().UTC(),
	})
	if len(b.records) > 500 {
		b.records = b.records[1:]
	}
}

func (b *MetricsBridge) Export() []MetricRecord {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.records
}
