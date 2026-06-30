package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/domain/habit"
)

type rowScanner interface {
	Scan(dest ...any) error
}

type HabitRepository struct {
	pool *pgxpool.Pool
}

func NewHabitRepository(pool *pgxpool.Pool) HabitRepository {
	return HabitRepository{pool: pool}
}

func (r HabitRepository) Create(ctx context.Context, h *habit.Habit) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO habits (id, user_id, name, icon, color, active_weekdays, monthly_target)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		h.ID, h.UserID, h.Name, h.Icon, h.Color, toInt16Slice(h.ActiveWeekdays), h.MonthlyTarget,
	)
	return err
}

func (r HabitRepository) FindByID(ctx context.Context, id uuid.UUID) (*habit.Habit, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, COALESCE(icon, ''), COALESCE(color, ''), active_weekdays, monthly_target, is_active, created_at, updated_at
		 FROM habits
		 WHERE id = $1 AND deleted_at IS NULL`,
		id,
	)

	h, err := scanHabit(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, habit.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (r HabitRepository) ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*habit.Habit, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, COALESCE(icon, ''), COALESCE(color, ''), active_weekdays, monthly_target, is_active, created_at, updated_at
		 FROM habits
		 WHERE user_id = $1 AND is_active AND deleted_at IS NULL
		 ORDER BY created_at`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []*habit.Habit
	for rows.Next() {
		h, err := scanHabit(rows)
		if err != nil {
			return nil, err
		}
		habits = append(habits, h)
	}

	return habits, rows.Err()
}

func (r HabitRepository) Update(ctx context.Context, h *habit.Habit) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE habits SET name = $1, icon = $2, color = $3, active_weekdays = $4, monthly_target = $5, updated_at = now()
		 WHERE id = $6`,
		h.Name, h.Icon, h.Color, toInt16Slice(h.ActiveWeekdays), h.MonthlyTarget, h.ID,
	)
	return err
}

func (r HabitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE habits SET deleted_at = now() WHERE id = $1`, id)
	return err
}

func scanHabit(row rowScanner) (*habit.Habit, error) {
	var h habit.Habit
	var weekdays []int16

	err := row.Scan(&h.ID, &h.UserID, &h.Name, &h.Icon, &h.Color, &weekdays, &h.MonthlyTarget, &h.IsActive, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		return nil, err
	}

	h.ActiveWeekdays = toIntSlice(weekdays)
	return &h, nil
}

func toInt16Slice(values []int) []int16 {
	result := make([]int16, len(values))
	for i, v := range values {
		result[i] = int16(v)
	}
	return result
}

func toIntSlice(values []int16) []int {
	result := make([]int, len(values))
	for i, v := range values {
		result[i] = int(v)
	}
	return result
}
