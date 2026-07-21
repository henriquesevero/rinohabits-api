package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/book"
)

type BookRepository interface {
	Create(ctx context.Context, b *book.Book) error
	FindByID(ctx context.Context, id uuid.UUID) (*book.Book, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*book.Book, error)
	ListByUserAndStatus(ctx context.Context, userID uuid.UUID, status book.Status) ([]*book.Book, error)
	Update(ctx context.Context, b *book.Book) error
	UpdateCover(ctx context.Context, id uuid.UUID, coverURL string) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllByUser(ctx context.Context, userID uuid.UUID) error
	ReorderBooks(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error
}
