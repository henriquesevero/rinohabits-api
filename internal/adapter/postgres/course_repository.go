package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/domain/course"
)

type CourseRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return CourseRepository{pool: pool}
}

func (r CourseRepository) Create(ctx context.Context, c *course.Course) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO courses (id, user_id, title, description, link, status, total_hours, current_hours, cover_url, started_at, finished_at, collection, sort_order)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
		   (SELECT COALESCE(MIN(sort_order), 0) - 1 FROM courses WHERE user_id = $2))`,
		c.ID, c.UserID, c.Title, c.Description, c.Link, string(c.Status), c.TotalHours, c.CurrentHours, c.CoverURL, c.StartedAt, c.FinishedAt, c.Collection,
	)
	return err
}

func (r CourseRepository) FindByID(ctx context.Context, id uuid.UUID) (*course.Course, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, description, link, status, total_hours, current_hours, sort_order, collection, cover_url, started_at, finished_at, created_at, updated_at
		 FROM courses WHERE id = $1`, id)
	c, err := scanCourse(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, course.ErrNotFound
	}
	return c, err
}

func (r CourseRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*course.Course, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, description, link, status, total_hours, current_hours, sort_order, collection, cover_url, started_at, finished_at, created_at, updated_at
		 FROM courses WHERE user_id = $1 ORDER BY sort_order ASC, created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectCourses(rows)
}

func (r CourseRepository) ListByUserAndStatus(ctx context.Context, userID uuid.UUID, status course.Status) ([]*course.Course, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, description, link, status, total_hours, current_hours, sort_order, collection, cover_url, started_at, finished_at, created_at, updated_at
		 FROM courses WHERE user_id = $1 AND status = $2 ORDER BY sort_order ASC, created_at DESC`, userID, string(status))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectCourses(rows)
}

func (r CourseRepository) Update(ctx context.Context, c *course.Course) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE courses SET title=$2, description=$3, link=$4, status=$5, total_hours=$6, current_hours=$7, cover_url=$8, started_at=$9, finished_at=$10, collection=$11
		 WHERE id=$1`,
		c.ID, c.Title, c.Description, c.Link, string(c.Status), c.TotalHours, c.CurrentHours, c.CoverURL, c.StartedAt, c.FinishedAt, c.Collection,
	)
	return err
}

func (r CourseRepository) UpdateCover(ctx context.Context, id uuid.UUID, coverURL string) error {
	_, err := r.pool.Exec(ctx, `UPDATE courses SET cover_url = $2 WHERE id = $1`, id, coverURL)
	return err
}

func (r CourseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM courses WHERE id = $1`, id)
	return err
}

func (r CourseRepository) ReorderCourses(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error {
	batch := &pgx.Batch{}
	for i, id := range ids {
		batch.Queue(`UPDATE courses SET sort_order = $1 WHERE id = $2 AND user_id = $3`, i, id, userID)
	}
	return r.pool.SendBatch(ctx, batch).Close()
}

func collectCourses(rows pgx.Rows) ([]*course.Course, error) {
	var courses []*course.Course
	for rows.Next() {
		c, err := scanCourse(rows)
		if err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

func scanCourse(row rowScanner) (*course.Course, error) {
	var c course.Course
	var statusStr string
	err := row.Scan(
		&c.ID, &c.UserID, &c.Title, &c.Description, &c.Link, &statusStr,
		&c.TotalHours, &c.CurrentHours, &c.SortOrder, &c.Collection, &c.CoverURL,
		&c.StartedAt, &c.FinishedAt,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	c.Status = course.Status(statusStr)
	return &c, nil
}
