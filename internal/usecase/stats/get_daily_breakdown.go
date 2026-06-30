package stats

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type DailyStatus struct {
	Date           time.Time
	RequiredCount  int
	CompletedCount int
	Percentage     float64
}

type GetDailyBreakdownUseCase struct {
	users  port.UserRepository
	habits port.HabitRepository
	logs   port.DailyLogRepository
	clock  port.Clock
}

func NewGetDailyBreakdownUseCase(users port.UserRepository, habits port.HabitRepository, logs port.DailyLogRepository, clock port.Clock) GetDailyBreakdownUseCase {
	return GetDailyBreakdownUseCase{users: users, habits: habits, logs: logs, clock: clock}
}

func (uc GetDailyBreakdownUseCase) Execute(ctx context.Context, userID uuid.UUID, periodType PeriodType, offset int) ([]DailyStatus, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	today, err := usecasehabit.LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return nil, err
	}

	start, end, err := periodRange(today, periodType, offset)
	if err != nil {
		return nil, err
	}

	allHabits, err := uc.habits.ListActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	results := make([]DailyStatus, 0, 31)
	for cursor := start; !cursor.After(end); cursor = cursor.AddDate(0, 0, 1) {
		breakdown, err := computeDayBreakdown(ctx, uc.logs, userID, allHabits, cursor, u.Timezone)
		if err != nil {
			return nil, err
		}

		results = append(results, DailyStatus{
			Date:           cursor,
			RequiredCount:  breakdown.requiredCount,
			CompletedCount: breakdown.completedCount,
			Percentage:     percentageOf(breakdown.completedCount, breakdown.requiredCount),
		})
	}

	return results, nil
}
