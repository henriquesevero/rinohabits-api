package habit

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type RestartHabitsUseCase struct {
	habits    port.HabitRepository
	dailyLogs port.DailyLogRepository
}

func NewRestartHabitsUseCase(habits port.HabitRepository, dailyLogs port.DailyLogRepository) RestartHabitsUseCase {
	return RestartHabitsUseCase{habits: habits, dailyLogs: dailyLogs}
}

// Execute clears completion history and moves every habit's creation date to
// now, so none of them count as required for any day before this moment —
// the habit definitions themselves (name, icon, schedule) are untouched.
func (uc RestartHabitsUseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	if err := uc.dailyLogs.DeleteAllByUser(ctx, userID); err != nil {
		return err
	}
	return uc.habits.RestartCreatedAt(ctx, userID)
}
