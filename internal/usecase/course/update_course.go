package course

import (
	"context"

	"github.com/google/uuid"

	domaincourse "github.com/henriquesevero/rinohabits-api/internal/domain/course"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
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

	// ScheduleProvided distinguishes "the request omitted schedule fields"
	// from "the request explicitly wants an empty schedule" — without it, a
	// plain title-only edit would wipe any existing schedule.
	ScheduleProvided bool
	ActiveWeekdays   []int
	ReminderHour     *int
	ReminderMinute   *int
}

type UpdateCourseUseCase struct {
	courses port.CourseRepository
	logs    port.CourseLogRepository
	users   port.UserRepository
	clock   port.Clock
}

func NewUpdateCourseUseCase(courses port.CourseRepository, logs port.CourseLogRepository, users port.UserRepository, clock port.Clock) UpdateCourseUseCase {
	return UpdateCourseUseCase{courses: courses, logs: logs, users: users, clock: clock}
}

func (uc UpdateCourseUseCase) Execute(ctx context.Context, in UpdateCourseInput) (c *domaincourse.Course, studiedToday bool, err error) {
	c, err = uc.courses.FindByID(ctx, in.CourseID)
	if err != nil {
		return nil, false, err
	}
	if c.UserID != in.UserID {
		return nil, false, domaincourse.ErrNotFound
	}

	now := uc.clock.Now()

	c.UpdateDetails(in.Title, in.Description, in.Link, in.TotalHours, in.Collection)

	if in.ScheduleProvided {
		if err := c.SetSchedule(in.ActiveWeekdays, in.ReminderHour, in.ReminderMinute); err != nil {
			return nil, false, err
		}
	}

	resetProgress := false
	if in.Status != "" && in.Status != c.Status {
		resetProgress, err = c.ChangeStatus(in.Status, now)
		if err != nil {
			return nil, false, err
		}
	}

	if err := uc.courses.Update(ctx, c); err != nil {
		return nil, false, err
	}
	c.UpdatedAt = now

	if resetProgress {
		if err := uc.logs.DeleteAllByCourse(ctx, c.ID); err != nil {
			return nil, false, err
		}
	}

	u, err := uc.users.FindByID(ctx, in.UserID)
	if err != nil {
		return nil, false, err
	}
	today, err := usecasehabit.LocalToday(now, u.Timezone)
	if err != nil {
		return nil, false, err
	}
	studiedToday, err = uc.logs.ExistsForCourseAndDate(ctx, c.ID, today)
	if err != nil {
		return nil, false, err
	}

	return c, studiedToday, nil
}
