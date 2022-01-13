package config

import "time"

type AuthConfig struct {
	Enable              bool          `mapstructure:"auth_enable"`
	SessionIDSize       int           `mapstructure:"session_id_size" validate:"gt=0"`
	SessionTTL          time.Duration `mapstructure:"session_ttl" validate:"gt=0"`
	SessionCookie       string        `mapstructure:"session_cookie" validate:"required"`
	SessionsRedisPrefix string        `mapstructure:"session_redis_prefix" validate:"required"`
	GoogleClientID      string        `mapstructure:"google_client_id"`
}
