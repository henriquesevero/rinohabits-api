package habit

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type TodayHabit struct {
	Habit           *domainhabit.Habit
	IsCompleted     bool
	WeekCompletions int
}

type ListTodayHabitsUseCase struct {
	habits port.HabitRepository
	logs   port.DailyLogRepository
	clock  port.Clock
	users  port.UserRepository
}

func NewListTodayHabitsUseCase(habits port.HabitRepository, logs port.DailyLogRepository, clock port.Clock, users port.UserRepository) ListTodayHabitsUseCase {
	return ListTodayHabitsUseCase{habits: habits, logs: logs, clock: clock, users: users}
}

func (uc ListTodayHabitsUseCase) Execute(ctx context.Context, userID uuid.UUID) ([]TodayHabit, time.Time, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, time.Time{}, err
	}

	today, err := LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return nil, time.Time{}, err
	}

	allHabits, err := uc.habits.ListActiveByUser(ctx, userID)
	if err != nil {
		return nil, time.Time{}, err
	}

	logsToday, err := uc.logs.ListByUserAndDate(ctx, userID, today)
	if err != nil {
		return nil, time.Time{}, err
	}

	completed := make(map[uuid.UUID]bool, len(logsToday))
	for _, l := range logsToday {
		completed[l.HabitID] = true
	}

	weekday := today.Weekday()

	// Only query week logs if there are frequency-based habits
	weekCount := make(map[uuid.UUID]int)
	for _, h := range allHabits {
		if h.IsFrequencyBased() {
			weekStart := isoWeekStart(today)
			weekEnd := weekStart.AddDate(0, 0, 7)
			weekLogs, err := uc.logs.ListByUserAndDateRange(ctx, userID, weekStart, weekEnd)
			if err != nil {
				return nil, time.Time{}, err
			}
			for _, l := range weekLogs {
				weekCount[l.HabitID]++
			}
			break
		}
	}

	result := make([]TodayHabit, 0, len(allHabits))
	for _, h := range allHabits {
		if h.IsFrequencyBased() {
			wc := weekCount[h.ID]
			if wc >= *h.WeeklyFrequency {
				continue // quota met — hide until next week
			}
			result = append(result, TodayHabit{
				Habit:           h,
				IsCompleted:     completed[h.ID],
				WeekCompletions: wc,
			})
		} else {
			if !h.IsRequiredOn(weekday) {
				continue
			}
			result = append(result, TodayHabit{
				Habit:       h,
				IsCompleted: completed[h.ID],
			})
		}
	}

	return result, today, nil
}

// isoWeekStart returns midnight UTC of the Monday that starts the ISO week containing t.
// t is expected to be midnight UTC of a local date (as returned by LocalToday).
func isoWeekStart(t time.Time) time.Time {
	weekday := t.Weekday()
	daysToMonday := int(weekday) - 1
	if weekday == time.Sunday {
		daysToMonday = 6
	}
	return t.AddDate(0, 0, -daysToMonday)
}
