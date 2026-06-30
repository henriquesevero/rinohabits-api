package habit

import (
	"context"

	"github.com/google/uuid"

	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type CreateHabitInput struct {
	UserID         uuid.UUID
	Name           string
	Icon           string
	Color          string
	ActiveWeekdays []int
	MonthlyTarget  *int
}

type CreateHabitUseCase struct {
	habits port.HabitRepository
}

func NewCreateHabitUseCase(habits port.HabitRepository) CreateHabitUseCase {
	return CreateHabitUseCase{habits: habits}
}

func (uc CreateHabitUseCase) Execute(ctx context.Context, in CreateHabitInput) (*domainhabit.Habit, error) {
	if len(in.ActiveWeekdays) == 0 {
		return nil, domainhabit.ErrNoActiveWeekday
	}

	h := domainhabit.New(in.UserID, in.Name, in.Icon, in.Color, in.ActiveWeekdays, in.MonthlyTarget)

	if err := uc.habits.Create(ctx, h); err != nil {
		return nil, err
	}

	return h, nil
}
