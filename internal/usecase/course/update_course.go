package course

import (
	"context"

	"github.com/google/uuid"

	domaincourse "github.com/henriquesevero/rinohabits-api/internal/domain/course"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type UpdateCourseInput struct {
	UserID      uuid.UUID
	CourseID    uuid.UUID
	Title       *string
	Description *string
	Link        *string
	TotalHours  *float64
	Status      domaincourse.Status
	Collection  *string
}

type UpdateCourseUseCase struct {
	courses port.CourseRepository
	clock   port.Clock
}

func NewUpdateCourseUseCase(courses port.CourseRepository, clock port.Clock) UpdateCourseUseCase {
	return UpdateCourseUseCase{courses: courses, clock: clock}
}

func (uc UpdateCourseUseCase) Execute(ctx context.Context, in UpdateCourseInput) (*domaincourse.Course, error) {
	c, err := uc.courses.FindByID(ctx, in.CourseID)
	if err != nil {
		return nil, err
	}
	if c.UserID != in.UserID {
		return nil, domaincourse.ErrNotFound
	}

	now := uc.clock.Now()

	c.UpdateDetails(in.Title, in.Description, in.Link, in.TotalHours, in.Collection)

	if in.Status != "" && in.Status != c.Status {
		if err := c.ChangeStatus(in.Status, now); err != nil {
			return nil, err
		}
	}

	if err := uc.courses.Update(ctx, c); err != nil {
		return nil, err
	}
	c.UpdatedAt = now
	return c, nil
}
