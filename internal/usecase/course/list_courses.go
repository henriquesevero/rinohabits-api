package course

import (
	"context"

	"github.com/google/uuid"

	domaincourse "github.com/henriquesevero/rinohabits-api/internal/domain/course"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type ListCoursesUseCase struct {
	courses port.CourseRepository
	logs    port.CourseLogRepository
	users   port.UserRepository
	clock   port.Clock
}

func NewListCoursesUseCase(courses port.CourseRepository, logs port.CourseLogRepository, users port.UserRepository, clock port.Clock) ListCoursesUseCase {
	return ListCoursesUseCase{courses: courses, logs: logs, users: users, clock: clock}
}

// Execute lists the user's courses alongside a studiedToday set, so the UI
// can highlight courses that already have a study session logged today
// (in the user's local timezone) without an extra round trip per course.
func (uc ListCoursesUseCase) Execute(ctx context.Context, userID uuid.UUID, status *domaincourse.Status) (courses []*domaincourse.Course, studiedToday map[uuid.UUID]bool, err error) {
	if status != nil {
		courses, err = uc.courses.ListByUserAndStatus(ctx, userID, *status)
	} else {
		courses, err = uc.courses.ListByUser(ctx, userID)
	}
	if err != nil {
		return nil, nil, err
	}

	u, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	today, err := usecasehabit.LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return nil, nil, err
	}

	loggedIDs, err := uc.logs.ListLoggedCourseIDsForDate(ctx, userID, today)
	if err != nil {
		return nil, nil, err
	}
	studiedToday = make(map[uuid.UUID]bool, len(loggedIDs))
	for _, id := range loggedIDs {
		studiedToday[id] = true
	}

	return courses, studiedToday, nil
}
