package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type DeleteAccountInput struct {
	UserID          uuid.UUID
	CurrentPassword string
}

type DeleteAccountUseCase struct {
	users  port.UserRepository
	hasher port.PasswordHasher
}

func NewDeleteAccountUseCase(users port.UserRepository, hasher port.PasswordHasher) DeleteAccountUseCase {
	return DeleteAccountUseCase{users: users, hasher: hasher}
}

func (uc DeleteAccountUseCase) Execute(ctx context.Context, in DeleteAccountInput) error {
	u, err := uc.users.FindByID(ctx, in.UserID)
	if err != nil {
		return err
	}

	if !uc.hasher.Matches(in.CurrentPassword, u.PasswordHash) {
		return user.ErrWrongPassword
	}

	return uc.users.Delete(ctx, in.UserID)
}
