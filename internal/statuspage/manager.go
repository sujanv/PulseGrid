package statuspage

import (
	"sync"
	"time"
)

type ComponentStatus string

const (
	StatusOperational      ComponentStatus = "OPERATIONAL"
	StatusDegraded         ComponentStatus = "DEGRADED_PERFORMANCE"
	StatusPartialOutage    ComponentStatus = "PARTIAL_OUTAGE"
	StatusMajorOutage      ComponentStatus = "MAJOR_OUTAGE"
)

type Component struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Group       string          `json:"group"`
	Status      ComponentStatus `json:"status"`
	TargetID    string          `json:"target_id,omitempty"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Subscriber struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "email" or "webhook"
	Endpoint  string    `json:"endpoint"`
	CreatedAt time.Time `json:"created_at"`
}

type StatusPageManager struct {
	mu          sync.RWMutex
	components  map[string]*Component
	subscribers map[string]*Subscriber
}

func NewStatusPageManager() *StatusPageManager {
	return &StatusPageManager{
		components:  make(map[string]*Component),
		subscribers: make(map[string]*Subscriber),
	}
}

func (m *StatusPageManager) RegisterComponent(c *Component) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.UpdatedAt = time.Now().UTC()
	m.components[c.ID] = c
}

func (m *StatusPageManager) UpdateComponentStatus(id string, status ComponentStatus) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.components[id]; ok {
		c.Status = status
		c.UpdatedAt = time.Now().UTC()
		return true
	}
	return false
}

func (m *StatusPageManager) ListComponents() []*Component {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*Component, 0, len(m.components))
	for _, c := range m.components {
		list = append(list, c)
	}
	return list
}

func (m *StatusPageManager) AddSubscriber(sub *Subscriber) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sub.CreatedAt = time.Now().UTC()
	m.subscribers[sub.ID] = sub
}

func (m *StatusPageManager) GetOverallStatus() ComponentStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hasMajor := false
	hasPartial := false
	hasDegraded := false

	for _, c := range m.components {
		switch c.Status {
		case StatusMajorOutage:
			hasMajor = true
		case StatusPartialOutage:
			hasPartial = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasMajor {
		return StatusMajorOutage
	}
	if hasPartial {
		return StatusPartialOutage
	}
	if hasDegraded {
		return StatusDegraded
	}
	return StatusOperational
}
