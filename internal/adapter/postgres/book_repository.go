package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henriquesevero/rinohabits-api/internal/domain/book"
)

type BookRepository struct {
	pool *pgxpool.Pool
}

func NewBookRepository(pool *pgxpool.Pool) BookRepository {
	return BookRepository{pool: pool}
}

func (r BookRepository) Create(ctx context.Context, b *book.Book) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO books (id, user_id, title, author, status, total_pages, current_page, cover_url, started_at, finished_at, sort_order)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		   (SELECT COALESCE(MAX(sort_order), -1) + 1 FROM books WHERE user_id = $2))`,
		b.ID, b.UserID, b.Title, b.Author, string(b.Status), b.TotalPages, b.CurrentPage, b.CoverURL, b.StartedAt, b.FinishedAt,
	)
	return err
}

func (r BookRepository) FindByID(ctx context.Context, id uuid.UUID) (*book.Book, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, author, status, total_pages, current_page, cover_url, started_at, finished_at, created_at, updated_at
		 FROM books WHERE id = $1`, id)
	b, err := scanBook(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, book.ErrNotFound
	}
	return b, err
}

func (r BookRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*book.Book, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, author, status, total_pages, current_page, cover_url, started_at, finished_at, created_at, updated_at
		 FROM books WHERE user_id = $1 ORDER BY sort_order ASC, created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectBooks(rows)
}

func (r BookRepository) ListByUserAndStatus(ctx context.Context, userID uuid.UUID, status book.Status) ([]*book.Book, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, title, author, status, total_pages, current_page, cover_url, started_at, finished_at, created_at, updated_at
		 FROM books WHERE user_id = $1 AND status = $2 ORDER BY sort_order ASC, created_at DESC`, userID, string(status))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectBooks(rows)
}

func (r BookRepository) Update(ctx context.Context, b *book.Book) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE books SET title=$2, author=$3, status=$4, total_pages=$5, current_page=$6, cover_url=$7, started_at=$8, finished_at=$9
		 WHERE id=$1`,
		b.ID, b.Title, b.Author, string(b.Status), b.TotalPages, b.CurrentPage, b.CoverURL, b.StartedAt, b.FinishedAt,
	)
	return err
}

func (r BookRepository) UpdateCover(ctx context.Context, id uuid.UUID, coverURL string) error {
	_, err := r.pool.Exec(ctx, `UPDATE books SET cover_url = $2 WHERE id = $1`, id, coverURL)
	return err
}

func (r BookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM books WHERE id = $1`, id)
	return err
}

func (r BookRepository) ReorderBooks(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error {
	batch := &pgx.Batch{}
	for i, id := range ids {
		batch.Queue(`UPDATE books SET sort_order = $1 WHERE id = $2 AND user_id = $3`, i, id, userID)
	}
	return r.pool.SendBatch(ctx, batch).Close()
}

func collectBooks(rows pgx.Rows) ([]*book.Book, error) {
	var books []*book.Book
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

func scanBook(row rowScanner) (*book.Book, error) {
	var b book.Book
	var statusStr string
	err := row.Scan(
		&b.ID, &b.UserID, &b.Title, &b.Author, &statusStr,
		&b.TotalPages, &b.CurrentPage, &b.CoverURL,
		&b.StartedAt, &b.FinishedAt,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	b.Status = book.Status(statusStr)
	return &b, nil
}
