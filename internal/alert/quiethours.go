package alert

import (
	"sync"
	"time"

	"github.com/sujanv/PulseGrid/internal/models"
)

type QuietHoursRule struct {
	ID        string `json:"id"`
	StartHour int    `json:"start_hour"` // e.g. 22 (10 PM)
	EndHour   int    `json:"end_hour"`   // e.g. 7 (7 AM)
	TimeZone  string `json:"timezone"`   // e.g. "America/New_York"
	AllowP0   bool   `json:"allow_p0"`   // If true, Critical P0 outages bypass quiet hours
}

type QuietHoursFilter struct {
	mu    sync.RWMutex
	rules []*QuietHoursRule
}

func NewQuietHoursFilter() *QuietHoursFilter {
	return &QuietHoursFilter{}
}

func (f *QuietHoursFilter) AddRule(rule *QuietHoursRule) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rules = append(f.rules, rule)
}

func (f *QuietHoursFilter) ShouldSuppress(event *models.AlertEvent) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, rule := range f.rules {
		loc, err := time.LoadLocation(rule.TimeZone)
		if err != nil {
			loc = time.UTC
		}

		t := event.Timestamp.In(loc)
		hour := t.Hour()

		inQuiet := false
		if rule.StartHour > rule.EndHour {
			// Crosses midnight (e.g. 22 to 7)
			inQuiet = (hour >= rule.StartHour || hour < rule.EndHour)
		} else {
			inQuiet = (hour >= rule.StartHour && hour < rule.EndHour)
		}

		if inQuiet {
			if rule.AllowP0 && event.EventType == models.AlertOutage {
				return false // Do not suppress critical outages
			}
			return true
		}
	}
	return false
}
