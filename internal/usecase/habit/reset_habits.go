package habit

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type ResetHabitsUseCase struct {
	habits port.HabitRepository
}

func NewResetHabitsUseCase(habits port.HabitRepository) ResetHabitsUseCase {
	return ResetHabitsUseCase{habits: habits}
}

func (uc ResetHabitsUseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	return uc.habits.DeleteAllByUser(ctx, userID)
}
