package dailylog

import (
	"time"

	"github.com/google/uuid"
)

type DailyLog struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	HabitID     uuid.UUID
	LogDate     time.Time
	CompletedAt time.Time
}

func New(userID, habitID uuid.UUID, logDate time.Time) *DailyLog {
	return &DailyLog{
		ID:      uuid.New(),
		UserID:  userID,
		HabitID: habitID,
		LogDate: logDate,
	}
}
