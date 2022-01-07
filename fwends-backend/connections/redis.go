package connections

import (
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

func OpenRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis_endpoint"),
		Password: viper.GetString("redis_password"),
		DB:       0,
	})
}
