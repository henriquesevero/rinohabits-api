package book

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusWantToRead Status = "quero_ler"
	StatusReading    Status = "lendo"
	StatusRead       Status = "lido"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusWantToRead, StatusReading, StatusRead:
		return true
	default:
		return false
	}
}

type Book struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Author      string
	Status      Status
	TotalPages  *int
	CurrentPage int
	CoverURL    *string
	StartedAt   *time.Time
	FinishedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func New(userID uuid.UUID, title, author string, totalPages *int, status Status) *Book {
	return &Book{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       title,
		Author:      author,
		Status:      status,
		TotalPages:  totalPages,
		CurrentPage: 0,
	}
}
