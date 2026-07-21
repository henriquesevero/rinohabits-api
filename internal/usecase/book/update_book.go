package book

import (
	"context"

	"github.com/google/uuid"

	domainbook "github.com/henriquesevero/rinohabits-api/internal/domain/book"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type UpdateBookInput struct {
	UserID      uuid.UUID
	BookID      uuid.UUID
	Title       *string
	Author      *string
	TotalPages  *int
	Status      domainbook.Status
	CurrentPage *int
	Collection  *string
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

	now := uc.clock.Now()

	b.UpdateDetails(in.Title, in.Author, in.TotalPages, in.Collection)

	if in.Status != "" && in.Status != b.Status {
		if err := b.ChangeStatus(in.Status, now); err != nil {
			return nil, err
		}
	}

	if in.CurrentPage != nil {
		b.SetCurrentPage(*in.CurrentPage)
	}

	if err := uc.books.Update(ctx, b); err != nil {
		return nil, err
	}
	b.UpdatedAt = now
	return b, nil
}
