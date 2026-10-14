package streaming

import (
	"sync"
)

type EventType string

const (
	EventCheckResult EventType = "check:result"
	EventIncident    EventType = "incident:state"
	EventRemediation EventType = "remediation:triggered"
)

type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

type EventBus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]chan Event),
	}
}

func (eb *EventBus) Subscribe(topic EventType) <-chan Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	ch := make(chan Event, 50)
	eb.subscribers[topic] = append(eb.subscribers[topic], ch)
	return ch
}

func (eb *EventBus) Publish(e Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, ch := range eb.subscribers[e.Type] {
		select {
		case ch <- e:
		default:
		}
	}
}
