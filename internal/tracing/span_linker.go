package tracing

import (
	"sync"
	"time"
)

type SpanLink struct {
	TraceID     string    `json:"trace_id"`
	SpanID      string    `json:"span_id"`
	Reason      string    `json:"reason"` // e.g. "CASCADING_UPSTREAM_FAILURE"
	LinkedAt    time.Time `json:"linked_at"`
}

type DependencySpanChain struct {
	TargetID    string     `json:"target_id"`
	RootSpanID  string     `json:"root_span_id"`
	Links       []SpanLink `json:"links"`
}

type SpanLinker struct {
	mu     sync.RWMutex
	chains map[string]*DependencySpanChain
}

func NewSpanLinker() *SpanLinker {
	return &SpanLinker{
		chains: make(map[string]*DependencySpanChain),
	}
}

func (l *SpanLinker) LinkOutageSpan(targetID, traceID, spanID, reason string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	chain, exists := l.chains[targetID]
	if !exists {
		chain = &DependencySpanChain{
			TargetID:   targetID,
			RootSpanID: spanID,
			Links:      make([]SpanLink, 0),
		}
		l.chains[targetID] = chain
	}

	chain.Links = append(chain.Links, SpanLink{
		TraceID:  traceID,
		SpanID:   spanID,
		Reason:   reason,
		LinkedAt: time.Now().UTC(),
	})
}

func (l *SpanLinker) GetChain(targetID string) (*DependencySpanChain, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	c, ok := l.chains[targetID]
	return c, ok
}
