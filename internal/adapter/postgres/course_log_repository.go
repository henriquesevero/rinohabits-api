package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/domain/courselog"
)

type CourseLogRepository struct {
	pool *pgxpool.Pool
}

func NewCourseLogRepository(pool *pgxpool.Pool) CourseLogRepository {
	return CourseLogRepository{pool: pool}
}

func (r CourseLogRepository) Upsert(ctx context.Context, log *courselog.CourseLog) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO course_logs (id, user_id, course_id, log_date, hours_logged)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (course_id, log_date) DO UPDATE SET hours_logged = course_logs.hours_logged + excluded.hours_logged`,
		log.ID, log.UserID, log.CourseID, log.LogDate, log.HoursLogged,
	)
	return err
}

func (r CourseLogRepository) SumHoursByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (float64, error) {
	var total float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(hours_logged), 0) FROM course_logs
		 WHERE user_id = $1 AND log_date >= $2 AND log_date <= $3`,
		userID, start, end,
	).Scan(&total)
	return total, err
}

func (r CourseLogRepository) CountCoursesFinishedByUserAndDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time, timezone string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM courses
		 WHERE user_id = $1 AND status = 'concluido'
		 AND (finished_at AT TIME ZONE $4)::date >= $2::date AND (finished_at AT TIME ZONE $4)::date <= $3::date`,
		userID, start, end, timezone,
	).Scan(&count)
	return count, err
}
