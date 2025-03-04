package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
	"xrf197ilz35aq2/internal"
)

var (
	redisClient       *redis.Client
	once              sync.Once
	initializationErr error
)

func NewRedisClient(redisConfig internal.RedisConfig, ctx context.Context) (*redis.Client, error) {
	once.Do(func() {
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
			initializationErr = fmt.Errorf("redis connection error failur :: err=%w", err)
			return
		}
		redisClient = client
	})
	if initializationErr != nil {
		return nil, initializationErr
	}
	return redisClient, nil
}
