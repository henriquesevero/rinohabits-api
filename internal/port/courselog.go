package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/courselog"
)

type CourseLogRepository interface {
	Upsert(ctx context.Context, log *courselog.CourseLog) error
	SumHoursByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (float64, error)
	CountCoursesFinishedByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, timezone string) (int, error)
	DeleteAllByCourse(ctx context.Context, courseID uuid.UUID) error
}
