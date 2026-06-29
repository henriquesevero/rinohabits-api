package habit

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type TodayHabit struct {
	Habit       *domainhabit.Habit
	IsCompleted bool
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

	result := make([]TodayHabit, 0, len(allHabits))
	for _, h := range allHabits {
		if !h.IsRequiredOn(weekday) {
			continue
		}
		result = append(result, TodayHabit{Habit: h, IsCompleted: completed[h.ID]})
	}

	return result, today, nil
}
