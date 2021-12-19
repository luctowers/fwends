package connections

import (
	"database/sql"
	"fmt"
	"os"
)

func OpenPostgres() (*sql.DB, error) {
	// TODO: make postgres port configurable
	// TODO: add intraservice tls
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_DB"),
	)
	return sql.Open("postgres", connStr)
}
