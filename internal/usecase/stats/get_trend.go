package stats

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type TrendPoint struct {
	Label      string
	Percentage float64
}

type GetTrendUseCase struct {
	users  port.UserRepository
	habits port.HabitRepository
	logs   port.DailyLogRepository
	clock  port.Clock
}

func NewGetTrendUseCase(users port.UserRepository, habits port.HabitRepository, logs port.DailyLogRepository, clock port.Clock) GetTrendUseCase {
	return GetTrendUseCase{users: users, habits: habits, logs: logs, clock: clock}
}

func (uc GetTrendUseCase) Execute(ctx context.Context, userID uuid.UUID, periodType PeriodType, count int) ([]TrendPoint, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	today, err := usecasehabit.LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return nil, err
	}

	allHabits, err := uc.habits.ListActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	points := make([]TrendPoint, 0, count)

	for i := -(count - 1); i <= 0; i++ {
		start, end, err := periodRange(today, periodType, i)
		if err != nil {
			return nil, err
		}

		requiredTotal, completedTotal := 0, 0

		for cursor := start; !cursor.After(end); cursor = cursor.AddDate(0, 0, 1) {
			required := usecasehabit.RequiredHabitsOn(allHabits, cursor, u.Timezone)
			if len(required) == 0 {
				continue
			}

			logsOnDay, err := uc.logs.ListByUserAndDate(ctx, userID, cursor)
			if err != nil {
				return nil, err
			}

			requiredTotal += len(required)
			completedTotal += usecasehabit.CountCompleted(required, logsOnDay)
		}

		points = append(points, TrendPoint{
			Label:      formatPeriodLabel(periodType, start),
			Percentage: percentageOf(completedTotal, requiredTotal),
		})
	}

	return points, nil
}
