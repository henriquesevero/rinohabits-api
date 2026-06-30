package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	Timezone     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func New(name, email, passwordHash, timezone string) *User {
	return &User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Timezone:     timezone,
	}
}
