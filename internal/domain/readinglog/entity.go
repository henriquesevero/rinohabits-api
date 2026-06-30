package readinglog

import (
	"time"

	"github.com/google/uuid"
)

type ReadingLog struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	BookID    uuid.UUID
	LogDate   time.Time
	PagesRead int
	CreatedAt time.Time
}

func New(userID, bookID uuid.UUID, logDate time.Time, pagesRead int) *ReadingLog {
	return &ReadingLog{
		ID:        uuid.New(),
		UserID:    userID,
		BookID:    bookID,
		LogDate:   logDate,
		PagesRead: pagesRead,
	}
}
