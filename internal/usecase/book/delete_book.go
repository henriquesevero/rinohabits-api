package book

import (
	"context"

	"github.com/google/uuid"

	domainbook "github.com/henriquesevero/rinohabits-api/internal/domain/book"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type DeleteBookUseCase struct {
	books port.BookRepository
}

func NewDeleteBookUseCase(books port.BookRepository) DeleteBookUseCase {
	return DeleteBookUseCase{books: books}
}

func (uc DeleteBookUseCase) Execute(ctx context.Context, userID, bookID uuid.UUID) error {
	b, err := uc.books.FindByID(ctx, bookID)
	if err != nil {
		return err
	}
	if b.UserID != userID {
		return domainbook.ErrNotFound
	}
	return uc.books.Delete(ctx, bookID)
}
