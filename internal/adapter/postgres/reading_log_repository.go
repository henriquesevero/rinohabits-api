package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/domain/readinglog"
)

type ReadingLogRepository struct {
	pool *pgxpool.Pool
}

func NewReadingLogRepository(pool *pgxpool.Pool) ReadingLogRepository {
	return ReadingLogRepository{pool: pool}
}

func (r ReadingLogRepository) Upsert(ctx context.Context, log *readinglog.ReadingLog) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO reading_logs (id, user_id, book_id, log_date, pages_read)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (book_id, log_date) DO UPDATE SET pages_read = reading_logs.pages_read + excluded.pages_read`,
		log.ID, log.UserID, log.BookID, log.LogDate, log.PagesRead,
	)
	return err
}

func (r ReadingLogRepository) SumPagesByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error) {
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(pages_read), 0) FROM reading_logs
		 WHERE user_id = $1 AND log_date >= $2 AND log_date <= $3`,
		userID, start, end,
	).Scan(&total)
	return total, err
}

func (r ReadingLogRepository) CountBooksFinishedByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, timezone string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM books
		 WHERE user_id = $1 AND status = 'lido'
		 AND (finished_at AT TIME ZONE $4)::date >= $2::date AND (finished_at AT TIME ZONE $4)::date <= $3::date`,
		userID, start, end, timezone,
	).Scan(&count)
	return count, err
}
