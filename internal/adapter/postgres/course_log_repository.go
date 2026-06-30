package postgres

import (
	"context"

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
