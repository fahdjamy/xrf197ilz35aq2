package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"sync"
	"xrf197ilz35aq2/internal"
)

var (
	dbPool     *pgx.Conn
	pgInstance *Postgres
	pgOnce     sync.Once
	dbErr      error // Global variable to store initialization errors
)

type Postgres struct {
	Pool *pgx.Conn
}

func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.Pool.Ping(ctx)
}

func (pg *Postgres) Close(ctx context.Context) {
	_ = pg.Pool.Close(ctx) // ignoring error returned
}

func NewPGConnection(pgConfig internal.PostgresConfig, ctx context.Context) (pool *Postgres, err error) {
	pgOnce.Do(func() {
		dbPool, err = pgx.Connect(ctx, pgConfig.DatabaseURL)
		if err != nil {
			dbErr = fmt.Errorf("failed to connect to 'database': %w", err)
			return
		}

	})

	// Important: Check the global error variable *after* once.Do.
	if dbErr != nil {
		return nil, dbErr // Return the stored error
	}

	pgInstance = &Postgres{Pool: dbPool}

	return pgInstance, nil
}
