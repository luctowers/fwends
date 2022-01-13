package config

type PostgresConfig struct {
	Endpoint string `mapstructure:"postgres_endpoint" validate:"required"`
	User     string `mapstructure:"postgres_user" validate:"required"`
	Password string `mapstructure:"postgres_password" validate:"required"`
	DB       string `mapstructure:"postgres_db" validate:"required"`
	SSLMode  string `mapstructure:"postgres_ssl_mode" validate:"required"`
}
