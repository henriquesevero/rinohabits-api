package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type ChangeEmailInput struct {
	UserID          uuid.UUID
	CurrentPassword string
	NewEmail        string
}

type ChangeEmailUseCase struct {
	users  port.UserRepository
	hasher port.PasswordHasher
}

func NewChangeEmailUseCase(users port.UserRepository, hasher port.PasswordHasher) ChangeEmailUseCase {
	return ChangeEmailUseCase{users: users, hasher: hasher}
}

func (uc ChangeEmailUseCase) Execute(ctx context.Context, in ChangeEmailInput) error {
	u, err := uc.users.FindByID(ctx, in.UserID)
	if err != nil {
		return err
	}

	if !uc.hasher.Matches(in.CurrentPassword, u.PasswordHash) {
		return user.ErrWrongPassword
	}

	existing, err := uc.users.FindByEmail(ctx, in.NewEmail)
	if err == nil && existing.ID != in.UserID {
		return user.ErrEmailAlreadyRegistered
	}

	return uc.users.UpdateEmail(ctx, in.UserID, in.NewEmail)
}
