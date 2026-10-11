package alert

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sujanv/PulseGrid/internal/models"
)

type BatchGrouper struct {
	mu         sync.Mutex
	pending    map[string][]*models.AlertEvent
	window     time.Duration
	dispatcher *Dispatcher
}

func NewBatchGrouper(window time.Duration, d *Dispatcher) *BatchGrouper {
	if window <= 0 {
		window = 15 * time.Second
	}
	b := &BatchGrouper{
		pending:    make(map[string][]*models.AlertEvent),
		window:     window,
		dispatcher: d,
	}
	go b.flushLoop()
	return b
}

func (b *BatchGrouper) Enqueue(groupKey string, event *models.AlertEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.pending[groupKey] = append(b.pending[groupKey], event)
}

func (b *BatchGrouper) flushLoop() {
	ticker := time.NewTicker(b.window)
	defer ticker.Stop()

	for range ticker.C {
		b.mu.Lock()
		for key, events := range b.pending {
			if len(events) == 1 {
				b.dispatcher.Dispatch(events[0])
			} else if len(events) > 1 {
				var names []string
				for _, e := range events {
					names = append(names, e.TargetName)
				}
				summaryEvent := &models.AlertEvent{
					ID:        uuid.New().String(),
					TargetID:  "group-" + key,
					TargetName: fmt.Sprintf("Cluster Outage (%d services)", len(events)),
					EventType: models.AlertOutage,
					Message:   fmt.Sprintf("Multiple services down in cluster %s: %s", key, strings.Join(names, ", ")),
					Timestamp: time.Now().UTC(),
				}
				b.dispatcher.Dispatch(summaryEvent)
			}
		}
		b.pending = make(map[string][]*models.AlertEvent)
		b.mu.Unlock()
	}
}
