package course

import (
	"context"

	"github.com/google/uuid"

	domaincourse "github.com/henriquesevero/rinohabits-api/internal/domain/course"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type CreateCourseInput struct {
	UserID      uuid.UUID
	Title       string
	Description string
	Link        string
	TotalHours  *float64
	Status      domaincourse.Status
}

type CreateCourseUseCase struct {
	courses port.CourseRepository
}

func NewCreateCourseUseCase(courses port.CourseRepository) CreateCourseUseCase {
	return CreateCourseUseCase{courses: courses}
}

func (uc CreateCourseUseCase) Execute(ctx context.Context, in CreateCourseInput) (*domaincourse.Course, error) {
	if in.Title == "" {
		return nil, domaincourse.ErrInvalidTitle
	}

	status := in.Status
	if status == "" {
		status = domaincourse.StatusWantToTake
	} else if !status.IsValid() {
		return nil, domaincourse.ErrInvalidStatus
	}

	c := domaincourse.New(in.UserID, in.Title, in.Description, in.Link, in.TotalHours, status)
	if err := uc.courses.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}
