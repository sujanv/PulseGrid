package maintenance

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Window struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TargetIDs   []string  `json:"target_ids"` // Empty slice applies to ALL targets
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
}

func (w *Window) IsActiveAt(t time.Time) bool {
	return (t.Equal(w.StartTime) || t.After(w.StartTime)) && t.Before(w.EndTime)
}

func (w *Window) AppliesToTarget(targetID string) bool {
	if len(w.TargetIDs) == 0 {
		return true // global maintenance window
	}
	for _, id := range w.TargetIDs {
		if id == targetID {
			return true
		}
	}
	return false
}

type Manager struct {
	mu      sync.RWMutex
	windows map[string]*Window
}

func NewManager() *Manager {
	return &Manager{
		windows: make(map[string]*Window),
	}
}

func (m *Manager) AddWindow(title, desc string, targetIDs []string, start, end time.Time) *Window {
	m.mu.Lock()
	defer m.mu.Unlock()

	w := &Window{
		ID:          uuid.New().String(),
		Title:       title,
		Description: desc,
		TargetIDs:   targetIDs,
		StartTime:   start,
		EndTime:     end,
		CreatedAt:   time.Now().UTC(),
	}
	m.windows[w.ID] = w
	return w
}

func (m *Manager) RemoveWindow(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.windows[id]; ok {
		delete(m.windows, id)
		return true
	}
	return false
}

func (m *Manager) ListWindows() []*Window {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*Window, 0, len(m.windows))
	for _, w := range m.windows {
		list = append(list, w)
	}
	return list
}

func (m *Manager) IsUnderMaintenance(targetID string, at time.Time) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, w := range m.windows {
		if w.IsActiveAt(at) && w.AppliesToTarget(targetID) {
			return true
		}
	}
	return false
}
