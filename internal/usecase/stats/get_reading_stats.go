package stats

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type ReadingStats struct {
	PeriodType     PeriodType
	Offset         int
	Start          string
	End            string
	PagesRead      int
	BooksFinished  int
}

type GetReadingStatsUseCase struct {
	users       port.UserRepository
	readingLogs port.ReadingLogRepository
	clock       port.Clock
}

func NewGetReadingStatsUseCase(users port.UserRepository, readingLogs port.ReadingLogRepository, clock port.Clock) GetReadingStatsUseCase {
	return GetReadingStatsUseCase{users: users, readingLogs: readingLogs, clock: clock}
}

func (uc GetReadingStatsUseCase) Execute(ctx context.Context, userID uuid.UUID, periodType PeriodType, offset int) (ReadingStats, error) {
	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return ReadingStats{}, err
	}

	today, err := usecasehabit.LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return ReadingStats{}, err
	}

	start, end, err := periodRange(today, periodType, offset)
	if err != nil {
		return ReadingStats{}, err
	}

	pages, err := uc.readingLogs.SumPagesByUserAndDateRange(ctx, userID, start, end)
	if err != nil {
		return ReadingStats{}, err
	}

	books, err := uc.readingLogs.CountBooksFinishedByUserAndDateRange(ctx, userID, start, end)
	if err != nil {
		return ReadingStats{}, err
	}

	return ReadingStats{
		PeriodType:    periodType,
		Offset:        offset,
		Start:         start.Format("2006-01-02"),
		End:           end.Format("2006-01-02"),
		PagesRead:     pages,
		BooksFinished: books,
	}, nil
}
