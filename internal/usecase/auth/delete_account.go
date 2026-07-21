package auth

import (
	"context"
	"path"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type DeleteAccountInput struct {
	UserID          uuid.UUID
	CurrentPassword string
}

type DeleteAccountUseCase struct {
	users   port.UserRepository
	hasher  port.PasswordHasher
	books   port.BookRepository
	courses port.CourseRepository
	storage port.FileStorage
}

func NewDeleteAccountUseCase(
	users port.UserRepository,
	hasher port.PasswordHasher,
	books port.BookRepository,
	courses port.CourseRepository,
	storage port.FileStorage,
) DeleteAccountUseCase {
	return DeleteAccountUseCase{users: users, hasher: hasher, books: books, courses: courses, storage: storage}
}

func (uc DeleteAccountUseCase) Execute(ctx context.Context, in DeleteAccountInput) error {
	u, err := uc.users.FindByID(ctx, in.UserID)
	if err != nil {
		return err
	}

	if !uc.hasher.Matches(in.CurrentPassword, u.PasswordHash) {
		return user.ErrWrongPassword
	}

	uc.deleteStoredFiles(ctx, u)

	return uc.users.Delete(ctx, in.UserID)
}

// Best-effort: a storage hiccup must not block account deletion, since the
// database rows are the authoritative erasure record.
func (uc DeleteAccountUseCase) deleteStoredFiles(ctx context.Context, u *user.User) {
	if u.AvatarURL != nil {
		_ = uc.storage.Delete(ctx, "avatars/"+u.ID.String()+path.Ext(*u.AvatarURL))
	}

	if books, err := uc.books.ListByUser(ctx, u.ID); err == nil {
		for _, b := range books {
			if b.CoverURL != nil {
				_ = uc.storage.Delete(ctx, "books/"+b.ID.String()+path.Ext(*b.CoverURL))
			}
		}
	}

	if courses, err := uc.courses.ListByUser(ctx, u.ID); err == nil {
		for _, c := range courses {
			if c.CoverURL != nil {
				_ = uc.storage.Delete(ctx, "courses/"+c.ID.String()+path.Ext(*c.CoverURL))
			}
		}
	}
}
