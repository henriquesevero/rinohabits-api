package auth

import (
	"context"
	"errors"

	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type RegisterInput struct {
	Name       string
	Email      string
	Password   string
	Timezone   string
	InviteCode string
}

type RegisterUseCase struct {
	users              port.UserRepository
	hasher             port.PasswordHasher
	requiredInviteCode string
}

func NewRegisterUseCase(users port.UserRepository, hasher port.PasswordHasher, requiredInviteCode string) RegisterUseCase {
	return RegisterUseCase{users: users, hasher: hasher, requiredInviteCode: requiredInviteCode}
}

func (uc RegisterUseCase) Execute(ctx context.Context, in RegisterInput) (*user.User, error) {
	if uc.requiredInviteCode != "" && in.InviteCode != uc.requiredInviteCode {
		return nil, user.ErrInvalidInviteCode
	}

	existing, err := uc.users.FindByEmail(ctx, in.Email)
	if err != nil && !errors.Is(err, user.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, user.ErrEmailAlreadyRegistered
	}

	hash, err := uc.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}

	newUser := user.New(in.Name, in.Email, hash, in.Timezone)

	if err := uc.users.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}
