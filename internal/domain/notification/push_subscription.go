package notification

import (
	"time"

	"github.com/google/uuid"
)

type PushSubscription struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Endpoint       string
	P256DH         string
	Auth           string
	ReminderHour   int
	ReminderMinute int
	CreatedAt      time.Time
}

type ReminderTarget struct {
	Endpoint     string
	P256DH       string
	Auth         string
	Incomplete   int
}
