package port

import "github.com/google/uuid"

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Matches(plain, hash string) bool
}

type TokenManager interface {
	Generate(userID uuid.UUID) (string, error)
	Verify(token string) (uuid.UUID, error)
}
