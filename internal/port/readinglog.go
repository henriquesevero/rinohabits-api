package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/readinglog"
)

type MonthlyPages struct {
	Year  int
	Month time.Month
	Pages int
}

type ReadingLogRepository interface {
	Upsert(ctx context.Context, log *readinglog.ReadingLog) error
	SumPagesByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error)
	SumAllPagesByUser(ctx context.Context, userID uuid.UUID) (int, error)
	ListMonthlyPagesByUser(ctx context.Context, userID uuid.UUID) ([]MonthlyPages, error)
	CountBooksFinishedByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, timezone string) (int, error)
}
