package book

import (
	"context"
	"path"

	"github.com/google/uuid"

	domainbook "github.com/henriquesevero/rinohabits-api/internal/domain/book"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type DeleteBookUseCase struct {
	books   port.BookRepository
	storage port.FileStorage
}

func NewDeleteBookUseCase(books port.BookRepository, storage port.FileStorage) DeleteBookUseCase {
	return DeleteBookUseCase{books: books, storage: storage}
}

func (uc DeleteBookUseCase) Execute(ctx context.Context, userID, bookID uuid.UUID) error {
	b, err := uc.books.FindByID(ctx, bookID)
	if err != nil {
		return err
	}
	if b.UserID != userID {
		return domainbook.ErrNotFound
	}

	if b.CoverURL != nil {
		_ = uc.storage.Delete(ctx, "books/"+bookID.String()+path.Ext(*b.CoverURL))
	}

	return uc.books.Delete(ctx, bookID)
}
