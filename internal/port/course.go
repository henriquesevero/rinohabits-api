package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/course"
)

type CourseRepository interface {
	Create(ctx context.Context, c *course.Course) error
	FindByID(ctx context.Context, id uuid.UUID) (*course.Course, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*course.Course, error)
	ListByUserAndStatus(ctx context.Context, userID uuid.UUID, status course.Status) ([]*course.Course, error)
	Update(ctx context.Context, c *course.Course) error
	UpdateCover(ctx context.Context, id uuid.UUID, coverURL string) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllByUser(ctx context.Context, userID uuid.UUID) error
	ReorderCourses(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error
}
