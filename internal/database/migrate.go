package database

import (
	"database/sql"
	"embed"
	"fmt"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(db *sql.DB) error {
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	for _, entry := range entries {
		content, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		// TODO: add clause to check if migration has run

		_, err = db.Exec(string(content))
		if err != nil {
			return fmt.Errorf("failed to execute %s: %w", entry.Name(), err)
		}

		fmt.Printf("applied: %s\n", entry.Name())
	}

	return nil
}

func Fresh(db *sql.DB) error {
	_, err := db.Exec("PRAGMA foreign_keys=OFF")
	if err != nil {
		return fmt.Errorf("failed to disable foreign keys: %w", err)
	}

	rows, err := db.Query(
		"SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'",
	)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, name)
	}

	for _, table := range tables {
		_, err := db.Exec("DROP TABLE IF EXISTS " + table)
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	_, err = db.Exec("PRAGMA foreign_keys=ON")
	if err != nil {
		return fmt.Errorf("failed to re-enable foreign keys: %w", err)
	}

	fmt.Println("dropped all tables")
	return nil
}
