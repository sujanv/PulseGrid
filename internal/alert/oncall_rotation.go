package alert

import (
	"fmt"
	"time"
)

type ShiftDuration string

const (
	ShiftDaily  ShiftDuration = "DAILY"
	ShiftWeekly ShiftDuration = "WEEKLY"
)

type OnCallEngineer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type OnCallSchedule struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Engineers   []OnCallEngineer `json:"engineers"`
	ShiftType   ShiftDuration    `json:"shift_type"`
	StartDate   time.Time        `json:"start_date"`
	SecondaryID string           `json:"secondary_id,omitempty"`
}

type OnCallRotationManager struct {
	schedules map[string]*OnCallSchedule
}

func NewOnCallRotationManager() *OnCallRotationManager {
	return &OnCallRotationManager{
		schedules: make(map[string]*OnCallSchedule),
	}
}

func (m *OnCallRotationManager) RegisterSchedule(sched *OnCallSchedule) {
	m.schedules[sched.ID] = sched
}

func (m *OnCallRotationManager) GetActiveOnCall(scheduleID string, at time.Time) (*OnCallEngineer, error) {
	sched, exists := m.schedules[scheduleID]
	if !exists {
		return nil, fmt.Errorf("schedule not found: %s", scheduleID)
	}
	if len(sched.Engineers) == 0 {
		return nil, fmt.Errorf("no engineers assigned to schedule: %s", scheduleID)
	}

	shiftHours := 24.0
	if sched.ShiftType == ShiftWeekly {
		shiftHours = 24.0 * 7.0
	}

	diffHours := at.Sub(sched.StartDate).Hours()
	if diffHours < 0 {
		diffHours = 0
	}

	shiftIndex := int(diffHours/shiftHours) % len(sched.Engineers)
	return &sched.Engineers[shiftIndex], nil
}
