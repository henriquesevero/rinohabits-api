package habit

import (
	"time"

	"github.com/google/uuid"
)

type Habit struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Name            string
	Icon            string
	Color           string
	ActiveWeekdays  []int
	WeeklyFrequency *int
	MonthlyTarget   *int
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func New(userID uuid.UUID, name, icon, color string, activeWeekdays []int, weeklyFrequency *int, monthlyTarget *int) (*Habit, error) {
	if weeklyFrequency != nil {
		activeWeekdays = []int{}
	} else if !validWeekdays(activeWeekdays) {
		return nil, ErrInvalidWeekday
	}

	return &Habit{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            name,
		Icon:            icon,
		Color:           color,
		ActiveWeekdays:  activeWeekdays,
		WeeklyFrequency: weeklyFrequency,
		MonthlyTarget:   monthlyTarget,
		IsActive:        true,
	}, nil
}

func (h *Habit) SetSchedule(activeWeekdays []int, weeklyFrequency *int) error {
	if weeklyFrequency != nil {
		activeWeekdays = []int{}
	} else if !validWeekdays(activeWeekdays) {
		return ErrInvalidWeekday
	}

	h.ActiveWeekdays = activeWeekdays
	h.WeeklyFrequency = weeklyFrequency
	return nil
}

func validWeekdays(weekdays []int) bool {
	for _, d := range weekdays {
		if d < 1 || d > 7 {
			return false
		}
	}
	return true
}

func (h *Habit) IsFrequencyBased() bool {
	return h.WeeklyFrequency != nil
}

func (h *Habit) IsRequiredOn(weekday time.Weekday) bool {
	for _, d := range h.ActiveWeekdays {
		if d == isoWeekday(weekday) {
			return true
		}
	}
	return false
}

func isoWeekday(weekday time.Weekday) int {
	if weekday == time.Sunday {
		return 7
	}
	return int(weekday)
}
