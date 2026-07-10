package stats

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type StudyStats struct {
	PeriodType      PeriodType
	Offset          int
	Start           string
	End             string
	HoursStudied    float64
	CoursesFinished int
}

type GetStudyStatsUseCase struct {
	users      port.UserRepository
	courseLogs port.CourseLogRepository
	clock      port.Clock
}

func NewGetStudyStatsUseCase(users port.UserRepository, courseLogs port.CourseLogRepository, clock port.Clock) GetStudyStatsUseCase {
	return GetStudyStatsUseCase{users: users, courseLogs: courseLogs, clock: clock}
}

func (uc GetStudyStatsUseCase) Execute(ctx context.Context, userID uuid.UUID, periodType PeriodType, offset int) (StudyStats, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return StudyStats{}, err
	}

	today, err := usecasehabit.LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return StudyStats{}, err
	}

	start, end, err := periodRange(today, periodType, offset)
	if err != nil {
		return StudyStats{}, err
	}

	hours, err := uc.courseLogs.SumHoursByUserAndDateRange(ctx, userID, start, end)
	if err != nil {
		return StudyStats{}, err
	}

	courses, err := uc.courseLogs.CountCoursesFinishedByUserAndDateRange(ctx, userID, start, end, u.Timezone)
	if err != nil {
		return StudyStats{}, err
	}

	return StudyStats{
		PeriodType:      periodType,
		Offset:          offset,
		Start:           start.Format("2006-01-02"),
		End:             end.Format("2006-01-02"),
		HoursStudied:    hours,
		CoursesFinished: courses,
	}, nil
}
