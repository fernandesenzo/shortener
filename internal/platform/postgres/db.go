package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func NewConnection(dbURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db connection: %w", err)
	}
	return db, nil
}
