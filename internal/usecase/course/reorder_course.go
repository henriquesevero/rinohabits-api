package course

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type ReorderCoursesUseCase struct {
	courses port.CourseRepository
}

func NewReorderCoursesUseCase(courses port.CourseRepository) ReorderCoursesUseCase {
	return ReorderCoursesUseCase{courses: courses}
}

func (uc ReorderCoursesUseCase) Execute(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error {
	return uc.courses.ReorderCourses(ctx, userID, ids)
}
