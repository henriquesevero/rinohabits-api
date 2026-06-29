package user

import "errors"

var (
	ErrEmailAlreadyRegistered = errors.New("email already registered")
	ErrNotFound               = errors.New("user not found")
	ErrInvalidCredentials     = errors.New("invalid credentials")
)
