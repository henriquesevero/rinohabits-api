package habit

import (
	"context"

	"github.com/google/uuid"

	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type UpdateHabitInput struct {
	UserID          uuid.UUID
	HabitID         uuid.UUID
	Name            string
	Icon            string
	Color           string
	ActiveWeekdays  []int
	WeeklyFrequency *int
	MonthlyTarget   *int
}

type UpdateHabitUseCase struct {
	habits port.HabitRepository
}

func NewUpdateHabitUseCase(habits port.HabitRepository) UpdateHabitUseCase {
	return UpdateHabitUseCase{habits: habits}
}

func (uc UpdateHabitUseCase) Execute(ctx context.Context, in UpdateHabitInput) (*domainhabit.Habit, error) {
	if in.WeeklyFrequency == nil && len(in.ActiveWeekdays) == 0 {
		return nil, domainhabit.ErrNoSchedule
	}

	h, err := uc.habits.FindByID(ctx, in.HabitID)
	if err != nil {
		return nil, err
	}
	if h.UserID != in.UserID {
		return nil, domainhabit.ErrNotFound
	}

	h.Name = in.Name
	h.Icon = in.Icon
	h.Color = in.Color
	h.MonthlyTarget = in.MonthlyTarget
	if err := h.SetSchedule(in.ActiveWeekdays, in.WeeklyFrequency); err != nil {
		return nil, err
	}

	if err := uc.habits.Update(ctx, h); err != nil {
		return nil, err
	}

	return h, nil
}
