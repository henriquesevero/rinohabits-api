package stats

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type DayStatus string

const (
	DayStatusPerfect DayStatus = "perfect"
	DayStatusFailed  DayStatus = "failed"
	DayStatusNeutral DayStatus = "neutral"
	DayStatusFuture  DayStatus = "future"
)

type CalendarDay struct {
	Date              time.Time
	Status            DayStatus
	RequiredCount     int
	CompletedCount    int
	CompletedHabitIDs []uuid.UUID
}

type CalendarSummary struct {
	Days        []CalendarDay
	ActiveDays  int
	PerfectDays int
	TotalChecks int
	TotalHabits int
	Habits      []*domainhabit.Habit
}

type GetCalendarUseCase struct {
	users  port.UserRepository
	habits port.HabitRepository
	logs   port.DailyLogRepository
	clock  port.Clock
}

func NewGetCalendarUseCase(users port.UserRepository, habits port.HabitRepository, logs port.DailyLogRepository, clock port.Clock) GetCalendarUseCase {
	return GetCalendarUseCase{users: users, habits: habits, logs: logs, clock: clock}
}

func (uc GetCalendarUseCase) Execute(ctx context.Context, userID uuid.UUID, year int, month time.Month) (CalendarSummary, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return CalendarSummary{}, err
	}

	today, err := usecasehabit.LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return CalendarSummary{}, err
	}

	allHabits, err := uc.habits.ListActiveByUser(ctx, userID)
	if err != nil {
		return CalendarSummary{}, err
	}

	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)

	summary := CalendarSummary{
		Days:        make([]CalendarDay, 0, 31),
		TotalHabits: len(allHabits),
		Habits:      allHabits,
	}

	for cursor := start; !cursor.After(end); cursor = cursor.AddDate(0, 0, 1) {
		if cursor.After(today) {
			summary.Days = append(summary.Days, CalendarDay{Date: cursor, Status: DayStatusFuture})
			continue
		}

		breakdown, err := computeDayBreakdown(ctx, uc.logs, userID, allHabits, cursor, u.Timezone)
		if err != nil {
			return CalendarSummary{}, err
		}

		if breakdown.requiredCount == 0 {
			summary.Days = append(summary.Days, CalendarDay{Date: cursor, Status: DayStatusNeutral})
			continue
		}

		status := DayStatusFailed
		if breakdown.completedCount == breakdown.requiredCount {
			status = DayStatusPerfect
			summary.PerfectDays++
		}
		summary.ActiveDays++
		summary.TotalChecks += breakdown.completedCount

		summary.Days = append(summary.Days, CalendarDay{
			Date:              cursor,
			Status:            status,
			RequiredCount:     breakdown.requiredCount,
			CompletedCount:    breakdown.completedCount,
			CompletedHabitIDs: breakdown.completedHabitIDs,
		})
	}

	return summary, nil
}
