package book

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainbook "github.com/henriquesevero/rinohabits-api/internal/domain/book"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type UpdateBookInput struct {
	UserID     uuid.UUID
	BookID     uuid.UUID
	Title      string
	Author     string
	TotalPages *int
	Status     domainbook.Status
}

type UpdateBookUseCase struct {
	books port.BookRepository
	clock port.Clock
}

func NewUpdateBookUseCase(books port.BookRepository, clock port.Clock) UpdateBookUseCase {
	return UpdateBookUseCase{books: books, clock: clock}
}

func (uc UpdateBookUseCase) Execute(ctx context.Context, in UpdateBookInput) (*domainbook.Book, error) {
	b, err := uc.books.FindByID(ctx, in.BookID)
	if err != nil {
		return nil, err
	}
	if b.UserID != in.UserID {
		return nil, domainbook.ErrNotFound
	}

	if in.Title != "" {
		b.Title = in.Title
	}
	b.Author = in.Author
	b.TotalPages = in.TotalPages

	if in.Status != "" && in.Status != b.Status {
		if !in.Status.IsValid() {
			return nil, domainbook.ErrInvalidStatus
		}
		now := uc.clock.Now()
		if in.Status == domainbook.StatusReading && b.StartedAt == nil {
			b.StartedAt = &now
		}
		if in.Status == domainbook.StatusRead && b.FinishedAt == nil {
			b.FinishedAt = &now
		}
		if in.Status == domainbook.StatusWantToRead {
			b.StartedAt = nil
			b.FinishedAt = nil
			b.CurrentPage = 0
		}
		b.Status = in.Status
	}

	if err := uc.books.Update(ctx, b); err != nil {
		return nil, err
	}
	b.UpdatedAt = time.Now()
	return b, nil
}
