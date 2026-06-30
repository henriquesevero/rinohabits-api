package stats

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type dayBreakdown struct {
	requiredCount     int
	completedCount    int
	completedHabitIDs []uuid.UUID
}

func computeDayBreakdown(
	ctx context.Context,
	logs port.DailyLogRepository,
	userID uuid.UUID,
	allHabits []*domainhabit.Habit,
	day time.Time,
	timezone string,
) (dayBreakdown, error) {
	required := usecasehabit.RequiredHabitsOn(allHabits, day, timezone)
	if len(required) == 0 {
		return dayBreakdown{}, nil
	}

	logsOnDay, err := logs.ListByUserAndDate(ctx, userID, day)
	if err != nil {
		return dayBreakdown{}, err
	}

	completedIDs := usecasehabit.CompletedHabitIDs(logsOnDay)
	completedCount := usecasehabit.CountCompleted(required, logsOnDay)

	ids := make([]uuid.UUID, 0, len(completedIDs))
	for _, h := range required {
		if completedIDs[h.ID] {
			ids = append(ids, h.ID)
		}
	}

	return dayBreakdown{requiredCount: len(required), completedCount: completedCount, completedHabitIDs: ids}, nil
}
