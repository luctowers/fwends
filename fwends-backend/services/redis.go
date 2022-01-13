package services

import (
	"fwends-backend/config"

	"github.com/go-redis/redis/v8"
)

func NewRedis(cfg *config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisEndpoint,
		Password: cfg.RedisPassword,
		DB:       0,
	})
}
