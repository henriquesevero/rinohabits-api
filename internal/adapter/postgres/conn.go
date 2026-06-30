package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool configures the connection pool to work behind a transaction-mode
// PgBouncer (e.g. Supabase's pooler on port 6543). That kind of pooler can
// hand the same physical backend connection to different logical sessions
// mid-stream, which breaks pgx's default automatic server-side prepared
// statement cache ("prepared statement already exists", SQLSTATE 42P05).
// Simple protocol mode avoids server-side prepares entirely, trading a
// small amount of per-query overhead for correctness under pooling.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
