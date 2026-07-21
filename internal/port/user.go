package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
)

type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	ListAll(ctx context.Context) ([]*user.User, error)
	UpdateTimezone(ctx context.Context, id uuid.UUID, timezone string) error
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL string) error
	UpdateBookCollectionOrder(ctx context.Context, id uuid.UUID, order []string) error
	UpdateCourseCollectionOrder(ctx context.Context, id uuid.UUID, order []string) error
	Delete(ctx context.Context, id uuid.UUID) error
}
