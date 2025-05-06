package storage

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"sync"
	"time"
	"xrf197ilz35aq2/internal"
)

var (
	dbPool                 *pgxpool.Pool
	pgInstance             *Postgres
	pgOnce                 sync.Once
	redisClient            *redis.Client
	redisOnce              sync.Once
	pgInitializationErr    error // Global variable to store initialization errors
	redisInitializationErr error
	tsInitializationErr    error
	timescaleOnce          sync.Once
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

func NewPGConnection(ctx context.Context, dbUrl string, log slog.Logger) (pool *Postgres, err error) {
	pgOnce.Do(func() {
		pgxPoolConfig, err := pgxpool.ParseConfig(dbUrl)
		pgxPoolConfig.MaxConns = 21
		pgxPoolConfig.BeforeConnect = func(ctx context.Context, config *pgx.ConnConfig) error {
			log.Info("Connecting to database", "url", dbUrl)
			return nil
		}
		if err != nil {
			pgInitializationErr = fmt.Errorf("pgxpool.ParseConfig failure: %w", err)
			return
		}
		dbPool, err = pgxpool.NewWithConfig(ctx, pgxPoolConfig)
		if err != nil {
			pgInitializationErr = fmt.Errorf("failed to connect to 'database': %w", err)
			return
		}
	})

	// Important: Check the global error variable *after* once.Do.
	if pgInitializationErr != nil {
		return nil, pgInitializationErr // Return the stored error
	}

	pgInstance = &Postgres{Pool: dbPool}

	return pgInstance, nil
}

func NewRedisClient(redisConfig internal.RedisConfig, ctx context.Context) (*redis.Client, error) {
	redisOnce.Do(func() {
		client := redis.NewClient(&redis.Options{
			Addr:         redisConfig.Address,
			Password:     redisConfig.Password,
			DB:           redisConfig.Database,
			PoolSize:     redisConfig.PoolSize,
			MaxRetries:   redisConfig.MaxRetries,
			MinIdleConns: redisConfig.MinIdleConns,
			DialTimeout:  redisConfig.DialTimeout * time.Second,
			ReadTimeout:  redisConfig.ReadTimeout * time.Second,
			WriteTimeout: redisConfig.WriteTimeout * time.Second,
		})
		_, err := client.Ping(ctx).Result()
		if err != nil {
			redisInitializationErr = fmt.Errorf("redis connection error failur :: err=%w", err)
			return
		}
		redisClient = client
	})

	if redisInitializationErr != nil {
		return nil, redisInitializationErr
	}

	return redisClient, nil
}

type TimescaleDB struct {
	Pool *pgxpool.Pool
}

func (ts *TimescaleDB) Ping(ctx context.Context) error {
	return ts.Pool.Ping(ctx)
}
