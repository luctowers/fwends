package services

import (
	"database/sql"
	"fmt"
	"fwends-backend/config"
)

func NewPostgres(cfg *config.PostgresConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Endpoint,
		cfg.DB,
		cfg.SSLMode,
	)
	return sql.Open("postgres", connStr)
}
