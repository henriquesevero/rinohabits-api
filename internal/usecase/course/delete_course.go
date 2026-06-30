package course

import (
	"context"

	"github.com/google/uuid"

	domaincourse "github.com/henriquesevero/rinohabits-api/internal/domain/course"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type DeleteCourseUseCase struct {
	courses port.CourseRepository
}

func NewDeleteCourseUseCase(courses port.CourseRepository) DeleteCourseUseCase {
	return DeleteCourseUseCase{courses: courses}
}

func (uc DeleteCourseUseCase) Execute(ctx context.Context, userID, courseID uuid.UUID) error {
	c, err := uc.courses.FindByID(ctx, courseID)
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return domaincourse.ErrNotFound
	}
	return uc.courses.Delete(ctx, courseID)
}
