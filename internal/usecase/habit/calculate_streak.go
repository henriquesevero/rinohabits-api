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
		required := RequiredHabitsOn(allHabits, cursor, u.Timezone)

		if len(required) > 0 {
			logsOnDay, err := uc.logs.ListByUserAndDate(ctx, userID, cursor)
			if err != nil {
				return 0, err
			}

			completedIDs := CompletedHabitIDs(logsOnDay)
			effective := EffectiveRequiredHabits(required, cursor, u.Timezone, completedIDs)

			if len(effective) > 0 {
				switch {
				case CountCompleted(effective, logsOnDay) == len(effective):
					streak++
				case !cursor.Equal(today):
					return streak, nil
				}
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

// RequiredHabitsOn returns habits that exist as of `day` (created on or
// before it) and are scheduled for that weekday. The creation day itself is
// included here; use EffectiveRequiredHabits to apply the "setup day" rule
// before treating a day as missed.
func RequiredHabitsOn(habits []*domainhabit.Habit, day time.Time, timezone string) []*domainhabit.Habit {
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

// EffectiveRequiredHabits drops habits whose creation day is `day` and that
// were not completed that day: a habit created late at night must not show
// as a failed day for the few remaining minutes it existed. If the user does
// complete it right away on the creation day, it still counts as a success.
func EffectiveRequiredHabits(required []*domainhabit.Habit, day time.Time, timezone string, completedIDs map[uuid.UUID]bool) []*domainhabit.Habit {
	effective := make([]*domainhabit.Habit, 0, len(required))

	for _, h := range required {
		createdDate, err := LocalToday(h.CreatedAt, timezone)
		if err == nil && createdDate.Equal(day) && !completedIDs[h.ID] {
			continue
		}
		effective = append(effective, h)
	}

	return effective
}

func CountCompleted(required []*domainhabit.Habit, logs []*dailylog.DailyLog) int {
	completed := CompletedHabitIDs(logs)

	count := 0
	for _, h := range required {
		if completed[h.ID] {
			count++
		}
	}

	return count
}

func CompletedHabitIDs(logs []*dailylog.DailyLog) map[uuid.UUID]bool {
	completed := make(map[uuid.UUID]bool, len(logs))
	for _, l := range logs {
		completed[l.HabitID] = true
	}
	return completed
}
