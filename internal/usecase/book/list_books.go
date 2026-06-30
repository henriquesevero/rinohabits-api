package book

import (
	"context"

	"github.com/google/uuid"

	domainbook "github.com/henriquesevero/rinohabits-api/internal/domain/book"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type ListBooksUseCase struct {
	books port.BookRepository
}

func NewListBooksUseCase(books port.BookRepository) ListBooksUseCase {
	return ListBooksUseCase{books: books}
}

func (uc ListBooksUseCase) Execute(ctx context.Context, userID uuid.UUID, status *domainbook.Status) ([]*domainbook.Book, error) {
	if status != nil {
		return uc.books.ListByUserAndStatus(ctx, userID, *status)
	}
	return uc.books.ListByUser(ctx, userID)
}
