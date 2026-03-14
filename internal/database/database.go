package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const DB_PATH = ".db/hp.db"

func NewDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", DB_PATH)

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
