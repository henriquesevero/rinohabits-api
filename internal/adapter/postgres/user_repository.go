package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return UserRepository{pool: pool}
}

func (r UserRepository) Create(ctx context.Context, u *user.User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id, name, email, password_hash, timezone)
		 VALUES ($1, $2, $3, $4, $5)`,
		u.ID, u.Name, u.Email, u.PasswordHash, u.Timezone,
	)
	return err
}

func (r UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	return r.scanOne(ctx,
		`SELECT id, name, email, password_hash, timezone, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	)
}

func (r UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return r.scanOne(ctx,
		`SELECT id, name, email, password_hash, timezone, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	)
}

func (r UserRepository) scanOne(ctx context.Context, query string, args ...any) (*user.User, error) {
	row := r.pool.QueryRow(ctx, query, args...)

	var u user.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Timezone, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, user.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}
