package auth

import (
	"context"
	"errors"

	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type LoginInput struct {
	Email    string
	Password string
}

type LoginUseCase struct {
	users  port.UserRepository
	hasher port.PasswordHasher
	tokens port.TokenManager
}

func NewLoginUseCase(users port.UserRepository, hasher port.PasswordHasher, tokens port.TokenManager) LoginUseCase {
	return LoginUseCase{users: users, hasher: hasher, tokens: tokens}
}

func (uc LoginUseCase) Execute(ctx context.Context, in LoginInput) (string, error) {
	u, err := uc.users.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return "", user.ErrInvalidCredentials
		}
		return "", err
	}

	if !uc.hasher.Matches(in.Password, u.PasswordHash) {
		return "", user.ErrInvalidCredentials
	}

	return uc.tokens.Generate(u.ID)
}
