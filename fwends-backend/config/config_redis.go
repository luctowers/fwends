package config

type RedisConfig struct {
	RedisEndpoint string `mapstructure:"redis_endpoint" validate:"required"`
	RedisPassword string `mapstructure:"redis_password" validate:"required"`
}
