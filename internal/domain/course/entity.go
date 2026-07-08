package course

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusShelf      Status = "na_prateleira"
	StatusWantToTake Status = "quero_fazer"
	StatusTaking     Status = "fazendo"
	StatusDone       Status = "concluido"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusShelf, StatusWantToTake, StatusTaking, StatusDone:
		return true
	default:
		return false
	}
}

type Course struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Title        string
	Description  string
	Link         string
	Status       Status
	TotalHours   *float64
	CurrentHours float64
	SortOrder    int
	Collection   *string
	CoverURL     *string
	StartedAt    *time.Time
	FinishedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func New(userID uuid.UUID, title, description, link string, totalHours *float64, status Status) *Course {
	return &Course{
		ID:           uuid.New(),
		UserID:       userID,
		Title:        title,
		Description:  description,
		Link:         link,
		Status:       status,
		TotalHours:   totalHours,
		CurrentHours: 0,
	}
}
