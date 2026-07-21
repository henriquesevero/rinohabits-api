package course

import (
	"context"
	"path"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type ResetCoursesUseCase struct {
	courses port.CourseRepository
	storage port.FileStorage
}

func NewResetCoursesUseCase(courses port.CourseRepository, storage port.FileStorage) ResetCoursesUseCase {
	return ResetCoursesUseCase{courses: courses, storage: storage}
}

func (uc ResetCoursesUseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	courses, err := uc.courses.ListByUser(ctx, userID)
	if err != nil {
		return err
	}

	for _, c := range courses {
		if c.CoverURL != nil {
			_ = uc.storage.Delete(ctx, "courses/"+c.ID.String()+path.Ext(*c.CoverURL))
		}
	}

	return uc.courses.DeleteAllByUser(ctx, userID)
}
