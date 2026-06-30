package course

import (
	"context"
	"time"

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

	if in.Title != nil && *in.Title != "" {
		c.Title = *in.Title
	}
	if in.Description != nil {
		c.Description = *in.Description
	}
	if in.Link != nil {
		c.Link = *in.Link
	}
	if in.TotalHours != nil {
		c.TotalHours = in.TotalHours
	}

	if in.Status != "" && in.Status != c.Status {
		if !in.Status.IsValid() {
			return nil, domaincourse.ErrInvalidStatus
		}
		now := uc.clock.Now()
		if in.Status == domaincourse.StatusTaking && c.StartedAt == nil {
			c.StartedAt = &now
		}
		if in.Status == domaincourse.StatusDone && c.FinishedAt == nil {
			c.FinishedAt = &now
		}
		if in.Status == domaincourse.StatusWantToTake {
			c.StartedAt = nil
			c.FinishedAt = nil
			c.CurrentHours = 0
		}
		c.Status = in.Status
	}

	if err := uc.courses.Update(ctx, c); err != nil {
		return nil, err
	}
	c.UpdatedAt = time.Now()
	return c, nil
}
