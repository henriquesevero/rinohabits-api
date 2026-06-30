package course

import (
	"context"

	"github.com/google/uuid"

	domaincourse "github.com/henriquesevero/rinohabits-api/internal/domain/course"
	"github.com/henriquesevero/rinohabits-api/internal/domain/courselog"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type RegisterStudyInput struct {
	UserID         uuid.UUID
	CourseID       uuid.UUID
	HoursLoggedNow float64
}

type RegisterStudyUseCase struct {
	courses port.CourseRepository
	logs    port.CourseLogRepository
	users   port.UserRepository
	clock   port.Clock
}

func NewRegisterStudyUseCase(courses port.CourseRepository, logs port.CourseLogRepository, users port.UserRepository, clock port.Clock) RegisterStudyUseCase {
	return RegisterStudyUseCase{courses: courses, logs: logs, users: users, clock: clock}
}

func (uc RegisterStudyUseCase) Execute(ctx context.Context, in RegisterStudyInput) (*domaincourse.Course, error) {
	if in.HoursLoggedNow <= 0 {
		return nil, domaincourse.ErrNoProgress
	}

	c, err := uc.courses.FindByID(ctx, in.CourseID)
	if err != nil {
		return nil, err
	}
	if c.UserID != in.UserID {
		return nil, domaincourse.ErrNotFound
	}

	u, err := uc.users.FindByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	today, err := usecasehabit.LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return nil, err
	}

	logEntry := courselog.New(in.UserID, in.CourseID, today, in.HoursLoggedNow)
	if err := uc.logs.Upsert(ctx, logEntry); err != nil {
		return nil, err
	}

	newHours := c.CurrentHours + in.HoursLoggedNow
	if c.TotalHours != nil && newHours > *c.TotalHours {
		newHours = *c.TotalHours
	}
	c.CurrentHours = newHours

	now := uc.clock.Now()
	if c.Status == domaincourse.StatusWantToTake {
		c.Status = domaincourse.StatusTaking
		c.StartedAt = &now
	}
	if c.TotalHours != nil && c.CurrentHours >= *c.TotalHours {
		c.Status = domaincourse.StatusDone
		c.FinishedAt = &now
	}

	if err := uc.courses.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}
