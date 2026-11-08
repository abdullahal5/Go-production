
package db

import (
    "context"
    "database/sql"
    "fmt"
    "time"
)

func Connect(databaseURL string) (*sql.DB, error) {
    database, err := sql.Open("pgx", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("sql.open: %w", err)
    }

    database.SetMaxOpenConns(25)
    database.SetMaxIdleConns(25)
    database.SetConnMaxLifetime(5 * time.Minute)

    ctx, cancel := context.WithTimeout(
        context.Background(),
        5*time.Second,
    )
    defer cancel()

    if err := database.PingContext(ctx); err != nil {
        _ = database.Close()
        return nil, fmt.Errorf("db.ping: %w", err)
    }

    return database, nil
}
