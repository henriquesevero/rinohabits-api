package stats

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type GetGamificationUseCase struct {
	users    port.UserRepository
	habits   port.HabitRepository
	logs     port.DailyLogRepository
	reading  port.ReadingLogRepository
	clock    port.Clock
}

func NewGetGamificationUseCase(
	users port.UserRepository,
	habits port.HabitRepository,
	logs port.DailyLogRepository,
	reading port.ReadingLogRepository,
	clock port.Clock,
) GetGamificationUseCase {
	return GetGamificationUseCase{users: users, habits: habits, logs: logs, reading: reading, clock: clock}
}

func (uc GetGamificationUseCase) Execute(ctx context.Context, userID uuid.UUID) (GamificationResult, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return GamificationResult{}, err
	}

	habits, err := uc.habits.ListActiveByUser(ctx, userID)
	if err != nil {
		return GamificationResult{}, err
	}

	allLogs, err := uc.logs.ListAllByUser(ctx, userID)
	if err != nil {
		return GamificationResult{}, err
	}

	monthlyPages, err := uc.reading.ListMonthlyPagesByUser(ctx, userID)
	if err != nil {
		return GamificationResult{}, err
	}

	streak, err := usecasehabit.NewCalculateStreakUseCase(uc.habits, uc.logs, uc.clock, uc.users).Execute(ctx, userID)
	if err != nil {
		return GamificationResult{}, err
	}

	result := computeGamification(habits, allLogs, monthlyPages, streak, u.Timezone)
	result.CurrentStreak = streak
	return result, nil
}
