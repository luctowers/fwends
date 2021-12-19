package connections

import (
	"os"

	"github.com/go-redis/redis/v8"
)

func OpenRedis() *redis.Client {
	// TODO: make redis port configurable
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":6379",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}
