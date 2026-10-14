package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Span struct {
	TraceID    string                 `json:"trace_id"`
	SpanID     string                 `json:"span_id"`
	ParentID   string                 `json:"parent_id,omitempty"`
	Name       string                 `json:"name"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	DurationMs float64                `json:"duration_ms"`
	Attributes map[string]interface{} `json:"attributes"`
	Events     []SpanEvent            `json:"events,omitempty"`
}

type SpanEvent struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

type Tracer struct {
	serviceName string
	mu          sync.RWMutex
	spans       map[string][]*Span // target_id -> spans
}

func NewTracer(serviceName string) *Tracer {
	return &Tracer{
		serviceName: serviceName,
		spans:       make(map[string][]*Span),
	}
}

func (t *Tracer) StartSpan(ctx context.Context, name string, parentID string) *Span {
	traceID := make([]byte, 16)
	spanID := make([]byte, 8)
	_, _ = rand.Read(traceID)
	_, _ = rand.Read(spanID)

	return &Span{
		TraceID:    hex.EncodeToString(traceID),
		SpanID:     hex.EncodeToString(spanID),
		ParentID:   parentID,
		Name:       name,
		StartTime:  time.Now().UTC(),
		Attributes: make(map[string]interface{}),
	}
}

func (s *Span) SetAttribute(key string, value interface{}) {
	s.Attributes[key] = value
}

func (s *Span) AddEvent(name string) {
	s.Events = append(s.Events, SpanEvent{
		Name:      name,
		Timestamp: time.Now().UTC(),
	})
}

func (s *Span) End() {
	s.EndTime = time.Now().UTC()
	s.DurationMs = float64(s.EndTime.Sub(s.StartTime).Microseconds()) / 1000.0
}

func (t *Tracer) RecordSpan(targetID string, s *Span) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans[targetID] = append(t.spans[targetID], s)
	if len(t.spans[targetID]) > 100 {
		t.spans[targetID] = t.spans[targetID][1:]
	}
}

func (t *Tracer) GetSpans(targetID string) []*Span {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.spans[targetID]
}
