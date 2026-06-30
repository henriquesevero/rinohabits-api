package auth

import (
	"context"
	"errors"
	"time"

	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type LoginInput struct {
	Email    string
	Password string
	Timezone string
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

	if in.Timezone != "" && in.Timezone != u.Timezone {
		if _, err := time.LoadLocation(in.Timezone); err == nil {
			_ = uc.users.UpdateTimezone(ctx, u.ID, in.Timezone)
		}
	}

	return uc.tokens.Generate(u.ID)
}
