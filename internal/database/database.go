package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func NewDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to set journal mode: %w", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return db, nil
}
