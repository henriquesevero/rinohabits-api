package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/dailylog"
	"github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
)

type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}

type HabitRepository interface {
	Create(ctx context.Context, h *habit.Habit) error
	FindByID(ctx context.Context, id uuid.UUID) (*habit.Habit, error)
	ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*habit.Habit, error)
}

type DailyLogRepository interface {
	Create(ctx context.Context, log *dailylog.DailyLog) error
	Delete(ctx context.Context, habitID uuid.UUID, logDate time.Time) error
	ListByUserAndDate(ctx context.Context, userID uuid.UUID, logDate time.Time) ([]*dailylog.DailyLog, error)
}
