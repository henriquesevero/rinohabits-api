package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/domain/dailylog"
)

type DailyLogRepository struct {
	pool *pgxpool.Pool
}

func NewDailyLogRepository(pool *pgxpool.Pool) DailyLogRepository {
	return DailyLogRepository{pool: pool}
}

func (r DailyLogRepository) Create(ctx context.Context, log *dailylog.DailyLog) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO daily_logs (id, user_id, habit_id, log_date)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (habit_id, log_date) DO NOTHING`,
		log.ID, log.UserID, log.HabitID, log.LogDate,
	)
	return err
}

func (r DailyLogRepository) Delete(ctx context.Context, habitID uuid.UUID, logDate time.Time) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM daily_logs WHERE habit_id = $1 AND log_date = $2`,
		habitID, logDate,
	)
	return err
}

func (r DailyLogRepository) ListByUserAndDate(ctx context.Context, userID uuid.UUID, logDate time.Time) ([]*dailylog.DailyLog, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, habit_id, log_date, completed_at
		 FROM daily_logs
		 WHERE user_id = $1 AND log_date = $2`,
		userID, logDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*dailylog.DailyLog
	for rows.Next() {
		var l dailylog.DailyLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.HabitID, &l.LogDate, &l.CompletedAt); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}

	return logs, rows.Err()
}
