package course

import (
	"context"

	"github.com/google/uuid"

	domaincourse "github.com/henriquesevero/rinohabits-api/internal/domain/course"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type ListCoursesUseCase struct {
	courses port.CourseRepository
}

func NewListCoursesUseCase(courses port.CourseRepository) ListCoursesUseCase {
	return ListCoursesUseCase{courses: courses}
}

func (uc ListCoursesUseCase) Execute(ctx context.Context, userID uuid.UUID, status *domaincourse.Status) ([]*domaincourse.Course, error) {
	if status != nil {
		return uc.courses.ListByUserAndStatus(ctx, userID, *status)
	}
	return uc.courses.ListByUser(ctx, userID)
}
