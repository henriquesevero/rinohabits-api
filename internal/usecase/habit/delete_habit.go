package habit

import (
	"context"

	"github.com/google/uuid"

	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type DeleteHabitUseCase struct {
	habits port.HabitRepository
}

func NewDeleteHabitUseCase(habits port.HabitRepository) DeleteHabitUseCase {
	return DeleteHabitUseCase{habits: habits}
}

func (uc DeleteHabitUseCase) Execute(ctx context.Context, userID, habitID uuid.UUID) error {
	h, err := uc.habits.FindByID(ctx, habitID)
	if err != nil {
		return err
	}
	if h.UserID != userID {
		return domainhabit.ErrNotFound
	}

	return uc.habits.Delete(ctx, habitID)
}
