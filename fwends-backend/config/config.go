package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// top-level
	HTTPDebug bool   `mapstructure:"http_debug"`
	HTTPPort  uint16 `mapstructure:"http_port" validate:"gt=0"`
	LogDebug  bool   `mapstructure:"log_debug"`

	// service configs
	Auth     AuthConfig     `mapstructure:",squash"`
	Postgres PostgresConfig `mapstructure:",squash"`
	Redis    RedisConfig    `mapstructure:",squash"`
	S3       S3Config       `mapstructure:",squash"`
}

func BindEnv(v *viper.Viper) {
	v.BindEnv("log_debug")
	v.BindEnv("http_port")
	v.BindEnv("http_debug")
	v.BindEnv("auth_enable")
	v.BindEnv("session_id_size")
	v.BindEnv("session_ttl")
	v.BindEnv("session_cookie")
	v.BindEnv("session_redis_prefix")
	v.BindEnv("google_client_id")
	v.BindEnv("postgres_endpoint")
	v.BindEnv("postgres_user")
	v.BindEnv("postgres_password")
	v.BindEnv("postgres_db")
	v.BindEnv("postgres_ssl_mode")
	v.BindEnv("redis_endpoint")
	v.BindEnv("redis_password")
	v.BindEnv("s3_endpoint")
	v.BindEnv("s3_region")
	v.BindEnv("s3_access_key")
	v.BindEnv("s3_secret_key")
	v.BindEnv("s3_media_bucket")
}

func SetDefaults(v *viper.Viper) {
	v.SetDefault("log_debug", false)
	v.SetDefault("http_port", 80)
	v.SetDefault("http_debug", false)
	v.SetDefault("auth_enable", true)
	v.SetDefault("session_id_size", 32)
	v.SetDefault("session_ttl", 24*time.Hour)
	v.SetDefault("session_cookie", "fwends_session")
	v.SetDefault("session_redis_prefix", "session/")
	v.SetDefault("postgres_ssl_mode", "require")
}
