package stats

import (
	"context"
	"sort"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type RankEntry struct {
	User   *user.User
	Result GamificationResult
	Rank   int
}

type GetRankingUseCase struct {
	users   port.UserRepository
	habits  port.HabitRepository
	logs    port.DailyLogRepository
	reading port.ReadingLogRepository
	clock   port.Clock
}

func NewGetRankingUseCase(
	users port.UserRepository,
	habits port.HabitRepository,
	logs port.DailyLogRepository,
	reading port.ReadingLogRepository,
	clock port.Clock,
) GetRankingUseCase {
	return GetRankingUseCase{users: users, habits: habits, logs: logs, reading: reading, clock: clock}
}

func (uc GetRankingUseCase) Execute(ctx context.Context, currentUserID uuid.UUID) ([]RankEntry, error) {
	allUsers, err := uc.users.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	streakUC := usecasehabit.NewCalculateStreakUseCase(uc.habits, uc.logs, uc.clock, uc.users)

	entries := make([]RankEntry, 0, len(allUsers))
	for _, u := range allUsers {
		habits, err := uc.habits.ListActiveByUser(ctx, u.ID)
		if err != nil {
			continue
		}

		allLogs, err := uc.logs.ListAllByUser(ctx, u.ID)
		if err != nil {
			continue
		}

		totalPages, err := uc.reading.SumAllPagesByUser(ctx, u.ID)
		if err != nil {
			continue
		}

		streak, err := streakUC.Execute(ctx, u.ID)
		if err != nil {
			streak = 0
		}

		result := computeGamification(habits, allLogs, totalPages, streak, u.Timezone)
		result.CurrentStreak = streak

		entries = append(entries, RankEntry{User: u, Result: result})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Result.TotalXP > entries[j].Result.TotalXP
	})

	for i := range entries {
		entries[i].Rank = i + 1
	}

	return entries, nil
}
