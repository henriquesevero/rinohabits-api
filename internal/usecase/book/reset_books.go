package book

import (
	"context"
	"path"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type ResetBooksUseCase struct {
	books   port.BookRepository
	storage port.FileStorage
}

func NewResetBooksUseCase(books port.BookRepository, storage port.FileStorage) ResetBooksUseCase {
	return ResetBooksUseCase{books: books, storage: storage}
}

func (uc ResetBooksUseCase) Execute(ctx context.Context, userID uuid.UUID) error {
	books, err := uc.books.ListByUser(ctx, userID)
	if err != nil {
		return err
	}

	for _, b := range books {
		if b.CoverURL != nil {
			_ = uc.storage.Delete(ctx, "books/"+b.ID.String()+path.Ext(*b.CoverURL))
		}
	}

	return uc.books.DeleteAllByUser(ctx, userID)
}
