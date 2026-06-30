package book

import (
	"context"

	"github.com/google/uuid"

	domainbook "github.com/henriquesevero/rinohabits-api/internal/domain/book"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type CreateBookInput struct {
	UserID     uuid.UUID
	Title      string
	Author     string
	TotalPages *int
	Status     domainbook.Status
}

type CreateBookUseCase struct {
	books port.BookRepository
}

func NewCreateBookUseCase(books port.BookRepository) CreateBookUseCase {
	return CreateBookUseCase{books: books}
}

func (uc CreateBookUseCase) Execute(ctx context.Context, in CreateBookInput) (*domainbook.Book, error) {
	if in.Title == "" {
		return nil, domainbook.ErrInvalidTitle
	}

	status := in.Status
	if status == "" {
		status = domainbook.StatusWantToRead
	} else if !status.IsValid() {
		return nil, domainbook.ErrInvalidStatus
	}

	b := domainbook.New(in.UserID, in.Title, in.Author, in.TotalPages, status)
	if err := uc.books.Create(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}
