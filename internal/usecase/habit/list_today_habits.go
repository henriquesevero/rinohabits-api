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

	weekCount, err := uc.weekCompletionCounts(ctx, userID, today, allHabits)
	if err != nil {
		return nil, time.Time{}, err
	}

	result := make([]TodayHabit, 0, len(allHabits))
	for _, h := range allHabits {
		if h.IsFrequencyBased() {
			wc := weekCount[h.ID]
			if wc >= *h.WeeklyFrequency {
				continue
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

func (uc ListTodayHabitsUseCase) weekCompletionCounts(ctx context.Context, userID uuid.UUID, today time.Time, allHabits []*domainhabit.Habit) (map[uuid.UUID]int, error) {
	weekCount := make(map[uuid.UUID]int)

	hasFrequencyBasedHabit := false
	for _, h := range allHabits {
		if h.IsFrequencyBased() {
			hasFrequencyBasedHabit = true
			break
		}
	}
	if !hasFrequencyBasedHabit {
		return weekCount, nil
	}

	weekStart := isoWeekStart(today)
	weekEnd := weekStart.AddDate(0, 0, 7)
	weekLogs, err := uc.logs.ListByUserAndDateRange(ctx, userID, weekStart, weekEnd)
	if err != nil {
		return nil, err
	}
	for _, l := range weekLogs {
		weekCount[l.HabitID]++
	}

	return weekCount, nil
}

// isoWeekStart returns midnight UTC of the Monday for the ISO week containing t.
// t must already be midnight UTC of a local date, as returned by LocalToday.
func isoWeekStart(t time.Time) time.Time {
	weekday := t.Weekday()
	daysToMonday := int(weekday) - 1
	if weekday == time.Sunday {
		daysToMonday = 6
	}
	return t.AddDate(0, 0, -daysToMonday)
}
