package course

import (
	"context"
	"path"

	"github.com/google/uuid"

	domaincourse "github.com/henriquesevero/rinohabits-api/internal/domain/course"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type DeleteCourseUseCase struct {
	courses port.CourseRepository
	storage port.FileStorage
}

func NewDeleteCourseUseCase(courses port.CourseRepository, storage port.FileStorage) DeleteCourseUseCase {
	return DeleteCourseUseCase{courses: courses, storage: storage}
}

func (uc DeleteCourseUseCase) Execute(ctx context.Context, userID, courseID uuid.UUID) error {
	c, err := uc.courses.FindByID(ctx, courseID)
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return domaincourse.ErrNotFound
	}

	if c.CoverURL != nil {
		_ = uc.storage.Delete(ctx, "courses/"+courseID.String()+path.Ext(*c.CoverURL))
	}

	return uc.courses.Delete(ctx, courseID)
}
