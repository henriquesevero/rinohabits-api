package book

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusOnShelf    Status = "na_estante"
	StatusWantToRead Status = "quero_ler"
	StatusReading    Status = "lendo"
	StatusRead       Status = "lido"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusOnShelf, StatusWantToRead, StatusReading, StatusRead:
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
	Collection  *string
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

func (b *Book) RegisterReading(pagesRead int, now time.Time) error {
	if pagesRead <= 0 {
		return ErrNoProgress
	}

	newPage := b.CurrentPage + pagesRead
	if b.TotalPages != nil && newPage > *b.TotalPages {
		newPage = *b.TotalPages
	}
	b.CurrentPage = newPage

	if b.Status == StatusWantToRead {
		b.Status = StatusReading
		b.StartedAt = &now
	}
	if b.TotalPages != nil && b.CurrentPage >= *b.TotalPages {
		b.Status = StatusRead
		b.FinishedAt = &now
	}

	return nil
}

func (b *Book) ChangeStatus(newStatus Status, now time.Time) error {
	if !newStatus.IsValid() {
		return ErrInvalidStatus
	}

	if newStatus == StatusReading && b.StartedAt == nil {
		b.StartedAt = &now
	}
	if newStatus == StatusReading {
		b.FinishedAt = nil
	}
	if newStatus == StatusRead && b.FinishedAt == nil {
		b.FinishedAt = &now
	}
	if newStatus == StatusOnShelf || newStatus == StatusWantToRead {
		b.StartedAt = nil
		b.FinishedAt = nil
		b.CurrentPage = 0
	}

	b.Status = newStatus
	return nil
}

func (b *Book) UpdateDetails(title, author *string, totalPages *int, collection *string) {
	if title != nil && *title != "" {
		b.Title = *title
	}
	if author != nil {
		b.Author = *author
	}
	if totalPages != nil {
		b.TotalPages = totalPages
	}
	if collection != nil {
		if *collection == "" {
			b.Collection = nil
		} else {
			b.Collection = collection
		}
	}
}

func (b *Book) SetCurrentPage(page int) {
	if page >= 0 {
		b.CurrentPage = page
	}
}
