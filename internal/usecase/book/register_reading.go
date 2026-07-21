package book

import (
	"context"

	"github.com/google/uuid"

	domainbook "github.com/henriquesevero/rinohabits-api/internal/domain/book"
	"github.com/henriquesevero/rinohabits-api/internal/domain/readinglog"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type RegisterReadingInput struct {
	UserID       uuid.UUID
	BookID       uuid.UUID
	PagesReadNow int
}

type RegisterReadingUseCase struct {
	books port.BookRepository
	logs  port.ReadingLogRepository
	users port.UserRepository
	clock port.Clock
}

func NewRegisterReadingUseCase(books port.BookRepository, logs port.ReadingLogRepository, users port.UserRepository, clock port.Clock) RegisterReadingUseCase {
	return RegisterReadingUseCase{books: books, logs: logs, users: users, clock: clock}
}

func (uc RegisterReadingUseCase) Execute(ctx context.Context, in RegisterReadingInput) (*domainbook.Book, error) {
	if in.PagesReadNow <= 0 {
		return nil, domainbook.ErrNoProgress
	}

	b, err := uc.books.FindByID(ctx, in.BookID)
	if err != nil {
		return nil, err
	}
	if b.UserID != in.UserID {
		return nil, domainbook.ErrNotFound
	}

	u, err := uc.users.FindByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}

	today, err := usecasehabit.LocalToday(uc.clock.Now(), u.Timezone)
	if err != nil {
		return nil, err
	}

	if err := b.RegisterReading(in.PagesReadNow, uc.clock.Now()); err != nil {
		return nil, err
	}

	logEntry := readinglog.New(in.UserID, in.BookID, today, in.PagesReadNow)
	if err := uc.logs.Upsert(ctx, logEntry); err != nil {
		return nil, err
	}

	if err := uc.books.Update(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}
