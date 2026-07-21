package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/dailylog"
)

type DailyLogRepository interface {
	Create(ctx context.Context, log *dailylog.DailyLog) error
	Delete(ctx context.Context, habitID uuid.UUID, logDate time.Time) error
	DeleteAllByUser(ctx context.Context, userID uuid.UUID) error
	ListByUserAndDate(ctx context.Context, userID uuid.UUID, logDate time.Time) ([]*dailylog.DailyLog, error)
	ListByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]*dailylog.DailyLog, error)
	ListAllByUser(ctx context.Context, userID uuid.UUID) ([]*dailylog.DailyLog, error)
}
