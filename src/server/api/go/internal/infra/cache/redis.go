package cache

import (
	"context"

	"github.com/memodb-io/Acontext/internal/config"
	"github.com/redis/go-redis/v9"
)

func New(cfg *config.Config) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	_ = rdb.Ping(context.Background()).Err()

	return rdb
}

func Close(rdb *redis.Client) error {
	return rdb.Close()
}
