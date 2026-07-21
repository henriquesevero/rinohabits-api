package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/habit"
)

type HabitRepository interface {
	Create(ctx context.Context, h *habit.Habit) error
	FindByID(ctx context.Context, id uuid.UUID) (*habit.Habit, error)
	ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*habit.Habit, error)
	Update(ctx context.Context, h *habit.Habit) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllByUser(ctx context.Context, userID uuid.UUID) error
	ReorderHabits(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error
}
