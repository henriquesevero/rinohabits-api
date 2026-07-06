package habit

import (
	"context"

	"github.com/google/uuid"

	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type CreateHabitInput struct {
	UserID          uuid.UUID
	Name            string
	Icon            string
	Color           string
	ActiveWeekdays  []int
	WeeklyFrequency *int
	MonthlyTarget   *int
}

type CreateHabitUseCase struct {
	habits port.HabitRepository
}

func NewCreateHabitUseCase(habits port.HabitRepository) CreateHabitUseCase {
	return CreateHabitUseCase{habits: habits}
}

func (uc CreateHabitUseCase) Execute(ctx context.Context, in CreateHabitInput) (*domainhabit.Habit, error) {
	if in.WeeklyFrequency == nil && len(in.ActiveWeekdays) == 0 {
		return nil, domainhabit.ErrNoSchedule
	}

	weekdays := in.ActiveWeekdays
	if in.WeeklyFrequency != nil {
		weekdays = []int{} // frequency habits have no specific weekdays
	}

	h := domainhabit.New(in.UserID, in.Name, in.Icon, in.Color, weekdays, in.WeeklyFrequency, in.MonthlyTarget)

	if err := uc.habits.Create(ctx, h); err != nil {
		return nil, err
	}

	return h, nil
}
