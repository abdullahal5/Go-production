package db

import (
	"database/sql"
	"fmt"
)

func Connect(databaseUrl string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseUrl)

	if err != nil {
		return nil, fmt.Errorf("sql.open: %w", err)
	}

	return db, nil
}
