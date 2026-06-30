package courselog

import (
	"time"

	"github.com/google/uuid"
)

type CourseLog struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	CourseID    uuid.UUID
	LogDate     time.Time
	HoursLogged float64
	CreatedAt   time.Time
}

func New(userID, courseID uuid.UUID, logDate time.Time, hoursLogged float64) *CourseLog {
	return &CourseLog{
		ID:          uuid.New(),
		UserID:      userID,
		CourseID:    courseID,
		LogDate:     logDate,
		HoursLogged: hoursLogged,
	}
}
