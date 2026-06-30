package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type ChangePasswordInput struct {
	UserID          uuid.UUID
	CurrentPassword string
	NewPassword     string
}

type ChangePasswordUseCase struct {
	users  port.UserRepository
	hasher port.PasswordHasher
}

func NewChangePasswordUseCase(users port.UserRepository, hasher port.PasswordHasher) ChangePasswordUseCase {
	return ChangePasswordUseCase{users: users, hasher: hasher}
}

func (uc ChangePasswordUseCase) Execute(ctx context.Context, in ChangePasswordInput) error {
	u, err := uc.users.FindByID(ctx, in.UserID)
	if err != nil {
		return err
	}

	if !uc.hasher.Matches(in.CurrentPassword, u.PasswordHash) {
		return user.ErrWrongPassword
	}

	newHash, err := uc.hasher.Hash(in.NewPassword)
	if err != nil {
		return err
	}

	return uc.users.UpdatePassword(ctx, in.UserID, newHash)
}
