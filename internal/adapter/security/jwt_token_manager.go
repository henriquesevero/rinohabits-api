package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

type JWTTokenManager struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTTokenManager(secret string, ttl time.Duration) JWTTokenManager {
	return JWTTokenManager{secret: []byte(secret), ttl: ttl}
}

type claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func (m JWTTokenManager) Generate(userID uuid.UUID) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	})
	return token.SignedString(m.secret)
}

func (m JWTTokenManager) Verify(tokenString string) (uuid.UUID, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &claims{}, func(*jwt.Token) (any, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil || !parsed.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	c, ok := parsed.Claims.(*claims)
	if !ok {
		return uuid.Nil, ErrInvalidToken
	}

	return c.UserID, nil
}
