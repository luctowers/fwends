package connections

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
)

func OpenPostgres() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=%s",
		viper.GetString("postgres_user"),
		viper.GetString("postgres_password"),
		viper.GetString("postgres_endpoint"),
		viper.GetString("postgres_db"),
		viper.GetString("postgres_ssl_mode"),
	)
	return sql.Open("postgres", connStr)
}
