package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/book"
	"github.com/henriquesevero/rinohabits-api/internal/domain/course"
	"github.com/henriquesevero/rinohabits-api/internal/domain/courselog"
	"github.com/henriquesevero/rinohabits-api/internal/domain/dailylog"
	"github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/domain/readinglog"
	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
)

type UserRepository interface {
	Create(ctx context.Context, u *user.User) error
	FindByEmail(ctx context.Context, email string) (*user.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	ListAll(ctx context.Context) ([]*user.User, error)
	UpdateTimezone(ctx context.Context, id uuid.UUID, timezone string) error
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type HabitRepository interface {
	Create(ctx context.Context, h *habit.Habit) error
	FindByID(ctx context.Context, id uuid.UUID) (*habit.Habit, error)
	ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*habit.Habit, error)
	Update(ctx context.Context, h *habit.Habit) error
	Delete(ctx context.Context, id uuid.UUID) error
	ReorderHabits(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error
}

type DailyLogRepository interface {
	Create(ctx context.Context, log *dailylog.DailyLog) error
	Delete(ctx context.Context, habitID uuid.UUID, logDate time.Time) error
	ListByUserAndDate(ctx context.Context, userID uuid.UUID, logDate time.Time) ([]*dailylog.DailyLog, error)
	ListByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]*dailylog.DailyLog, error)
	ListAllByUser(ctx context.Context, userID uuid.UUID) ([]*dailylog.DailyLog, error)
}

type BookRepository interface {
	Create(ctx context.Context, b *book.Book) error
	FindByID(ctx context.Context, id uuid.UUID) (*book.Book, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*book.Book, error)
	ListByUserAndStatus(ctx context.Context, userID uuid.UUID, status book.Status) ([]*book.Book, error)
	Update(ctx context.Context, b *book.Book) error
	UpdateCover(ctx context.Context, id uuid.UUID, coverURL string) error
	Delete(ctx context.Context, id uuid.UUID) error
	ReorderBooks(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error
}

type MonthlyPages struct {
	Year  int
	Month time.Month
	Pages int
}

type ReadingLogRepository interface {
	Upsert(ctx context.Context, log *readinglog.ReadingLog) error
	SumPagesByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error)
	SumAllPagesByUser(ctx context.Context, userID uuid.UUID) (int, error)
	ListMonthlyPagesByUser(ctx context.Context, userID uuid.UUID) ([]MonthlyPages, error)
	CountBooksFinishedByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, timezone string) (int, error)
}

type CourseRepository interface {
	Create(ctx context.Context, c *course.Course) error
	FindByID(ctx context.Context, id uuid.UUID) (*course.Course, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*course.Course, error)
	ListByUserAndStatus(ctx context.Context, userID uuid.UUID, status course.Status) ([]*course.Course, error)
	Update(ctx context.Context, c *course.Course) error
	UpdateCover(ctx context.Context, id uuid.UUID, coverURL string) error
	Delete(ctx context.Context, id uuid.UUID) error
	ReorderCourses(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error
}

type CourseLogRepository interface {
	Upsert(ctx context.Context, log *courselog.CourseLog) error
}
