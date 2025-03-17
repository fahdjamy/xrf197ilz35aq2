package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"sync"
	"xrf197ilz35aq2/internal"
)

var (
	dbPool     *pgxpool.Pool
	pgInstance *Postgres
	pgOnce     sync.Once
	dbErr      error // Global variable to store initialization errors
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func (pg *Postgres) Ping(ctx context.Context) error {
	return pg.Pool.Ping(ctx)
}

func (pg *Postgres) Close() {
	pg.Pool.Close() // ignoring error returned
}

func NewPGConnection(ctx context.Context, pgConfig internal.PostgresConfig, log slog.Logger) (pool *Postgres, err error) {
	pgOnce.Do(func() {
		pgxPoolConfig, err := pgxpool.ParseConfig(pgConfig.DatabaseURL)
		pgxPoolConfig.MaxConns = 21
		pgxPoolConfig.BeforeConnect = func(ctx context.Context, config *pgx.ConnConfig) error {
			log.Info("Connecting to database", "url", pgConfig.DatabaseURL)
			return nil
		}
		if err != nil {
			dbErr = fmt.Errorf("pgxpool.ParseConfig failure: %w", err)
			return
		}
		dbPool, err = pgxpool.NewWithConfig(ctx, pgxPoolConfig)
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
