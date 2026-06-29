package habit

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/dailylog"
	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type CalculateStreakUseCase struct {
	habits port.HabitRepository
	logs   port.DailyLogRepository
	clock  port.Clock
	users  port.UserRepository
}

func NewCalculateStreakUseCase(habits port.HabitRepository, logs port.DailyLogRepository, clock port.Clock, users port.UserRepository) CalculateStreakUseCase {
	return CalculateStreakUseCase{habits: habits, logs: logs, clock: clock, users: users}
}

func (uc CalculateStreakUseCase) Execute(ctx context.Context, userID uuid.UUID) (int, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return 0, err
	}

	allHabits, err := uc.habits.ListActiveByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	if len(allHabits) == 0 {
		return 0, nil
	}

	today, err := LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return 0, err
	}

	floor, err := earliestHabitDate(allHabits, today, u.Timezone)
	if err != nil {
		return 0, err
	}

	streak := 0
	cursor := today

	for !cursor.Before(floor) {
		required := requiredHabitsOn(allHabits, cursor, u.Timezone)

		if len(required) > 0 {
			logsOnDay, err := uc.logs.ListByUserAndDate(ctx, userID, cursor)
			if err != nil {
				return 0, err
			}

			switch {
			case allCompleted(required, logsOnDay):
				streak++
			case !cursor.Equal(today):
				return streak, nil
			}
		}

		cursor = cursor.AddDate(0, 0, -1)
	}

	return streak, nil
}

func earliestHabitDate(habits []*domainhabit.Habit, fallback time.Time, timezone string) (time.Time, error) {
	earliest := fallback

	for _, h := range habits {
		createdDate, err := LocalToday(h.CreatedAt, timezone)
		if err != nil {
			return time.Time{}, err
		}
		if createdDate.Before(earliest) {
			earliest = createdDate
		}
	}

	return earliest, nil
}

func requiredHabitsOn(habits []*domainhabit.Habit, day time.Time, timezone string) []*domainhabit.Habit {
	required := make([]*domainhabit.Habit, 0, len(habits))

	for _, h := range habits {
		createdDate, err := LocalToday(h.CreatedAt, timezone)
		if err != nil || createdDate.After(day) {
			continue
		}
		if h.IsRequiredOn(day.Weekday()) {
			required = append(required, h)
		}
	}

	return required
}

func allCompleted(required []*domainhabit.Habit, logs []*dailylog.DailyLog) bool {
	completed := make(map[uuid.UUID]bool, len(logs))
	for _, l := range logs {
		completed[l.HabitID] = true
	}

	for _, h := range required {
		if !completed[h.ID] {
			return false
		}
	}

	return true
}
