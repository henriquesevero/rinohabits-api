package habit

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/dailylog"
	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type ToggleHabitLogInput struct {
	UserID  uuid.UUID
	HabitID uuid.UUID
}

type ToggleHabitLogUseCase struct {
	habits port.HabitRepository
	logs   port.DailyLogRepository
	clock  port.Clock
	users  port.UserRepository
}

func NewToggleHabitLogUseCase(habits port.HabitRepository, logs port.DailyLogRepository, clock port.Clock, users port.UserRepository) ToggleHabitLogUseCase {
	return ToggleHabitLogUseCase{habits: habits, logs: logs, clock: clock, users: users}
}

func (uc ToggleHabitLogUseCase) Execute(ctx context.Context, in ToggleHabitLogInput) (bool, error) {
	h, err := uc.habits.FindByID(ctx, in.HabitID)
	if err != nil {
		return false, err
	}
	if h.UserID != in.UserID {
		return false, domainhabit.ErrNotFound
	}

	u, err := uc.users.FindByID(ctx, in.UserID)
	if err != nil {
		return false, err
	}

	today, err := LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return false, err
	}

	logsToday, err := uc.logs.ListByUserAndDate(ctx, in.UserID, today)
	if err != nil {
		return false, err
	}

	for _, l := range logsToday {
		if l.HabitID == h.ID {
			if err := uc.logs.Delete(ctx, h.ID, today); err != nil {
				return false, err
			}
			return false, nil
		}
	}

	if err := uc.logs.Create(ctx, dailylog.New(in.UserID, h.ID, today)); err != nil {
		return false, err
	}

	return true, nil
}
