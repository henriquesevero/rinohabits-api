package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
	"github.com/henriquesevero/rinohabits-api/internal/port"
)

type GetCurrentUserUseCase struct {
	users port.UserRepository
}

func NewGetCurrentUserUseCase(users port.UserRepository) GetCurrentUserUseCase {
	return GetCurrentUserUseCase{users: users}
}

func (uc GetCurrentUserUseCase) Execute(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return uc.users.FindByID(ctx, id)
}
