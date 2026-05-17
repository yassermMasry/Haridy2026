package cache

import (
	"context"
	"strconv"

	"haridy2026/configs"

	"github.com/redis/go-redis/v9"
)

func NewRedis(cfg configs.Config) *redis.Client {
	db, _ := strconv.Atoi(cfg.RedisDB)
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       db,
	})
}

func Ping(ctx context.Context, client *redis.Client) bool {
	return client.Ping(ctx).Err() == nil
}
