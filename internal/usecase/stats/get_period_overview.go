package stats

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type HabitProgress struct {
	Habit          *domainhabit.Habit
	RequiredCount  int
	CompletedCount int
	Percentage     float64
}

type PeriodOverview struct {
	PeriodType        PeriodType
	Offset            int
	Start             time.Time
	End               time.Time
	OverallPercentage float64
	RequiredTotal     int
	CompletedTotal    int
	Habits            []HabitProgress
}

type GetPeriodOverviewUseCase struct {
	users  port.UserRepository
	habits port.HabitRepository
	logs   port.DailyLogRepository
	clock  port.Clock
}

func NewGetPeriodOverviewUseCase(users port.UserRepository, habits port.HabitRepository, logs port.DailyLogRepository, clock port.Clock) GetPeriodOverviewUseCase {
	return GetPeriodOverviewUseCase{users: users, habits: habits, logs: logs, clock: clock}
}

func (uc GetPeriodOverviewUseCase) Execute(ctx context.Context, userID uuid.UUID, periodType PeriodType, offset int) (PeriodOverview, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return PeriodOverview{}, err
	}

	today, err := usecasehabit.LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return PeriodOverview{}, err
	}

	start, end, err := periodRange(today, periodType, offset)
	if err != nil {
		return PeriodOverview{}, err
	}

	allHabits, err := uc.habits.ListActiveByUser(ctx, userID)
	if err != nil {
		return PeriodOverview{}, err
	}

	progressByHabit := make(map[uuid.UUID]*HabitProgress, len(allHabits))
	for _, h := range allHabits {
		progressByHabit[h.ID] = &HabitProgress{Habit: h}
	}

	requiredTotal, completedTotal := 0, 0

	for cursor := start; !cursor.After(end); cursor = cursor.AddDate(0, 0, 1) {
		required := usecasehabit.RequiredHabitsOn(allHabits, cursor, u.Timezone)
		if len(required) == 0 {
			continue
		}

		logsOnDay, err := uc.logs.ListByUserAndDate(ctx, userID, cursor)
		if err != nil {
			return PeriodOverview{}, err
		}

		completedIDs := usecasehabit.CompletedHabitIDs(logsOnDay)
		effective := usecasehabit.EffectiveRequiredHabits(required, cursor, u.Timezone, completedIDs)

		for _, h := range effective {
			p := progressByHabit[h.ID]
			p.RequiredCount++
			requiredTotal++
			if completedIDs[h.ID] {
				p.CompletedCount++
				completedTotal++
			}
		}
	}

	habitsProgress := make([]HabitProgress, 0, len(allHabits))
	for _, h := range allHabits {
		p := progressByHabit[h.ID]

		denominator := p.RequiredCount
		if periodType == PeriodMonth && offset == 0 && h.MonthlyTarget != nil {
			denominator = *h.MonthlyTarget
		}

		if denominator == 0 {
			// no trackable days for this habit in the period yet (e.g. it was
			// just created): there is nothing to report progress on, so it
			// must not be displayed as "100% done".
			p.Percentage = 0
		} else {
			p.Percentage = percentageOf(p.CompletedCount, denominator)
		}
		habitsProgress = append(habitsProgress, *p)
	}

	overallPercentage := 0.0
	if requiredTotal > 0 {
		overallPercentage = percentageOf(completedTotal, requiredTotal)
	}

	return PeriodOverview{
		PeriodType:        periodType,
		Offset:            offset,
		Start:             start,
		End:               end,
		OverallPercentage: overallPercentage,
		RequiredTotal:     requiredTotal,
		CompletedTotal:    completedTotal,
		Habits:            habitsProgress,
	}, nil
}
