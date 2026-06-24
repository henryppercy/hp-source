package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

const schemaMigrationsDDL = `CREATE TABLE IF NOT EXISTS schema_migrations (
    version    TEXT PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT (datetime('now'))
)`

// Applies every migration in the embedded migrations directory not already
// recorded in schema_migrations. Returns the filenames it applied, in order.
func Migrate(db *sql.DB) ([]string, error) {
	migrations, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to open migrations: %w", err)
	}
	return migrateFS(db, migrations)
}

// Runs the migrations found in fsys (a flat directory of *.sql files) and
// returns the filenames it applied. Split out so tests can supply synthetic ones.
func migrateFS(db *sql.DB, fsys fs.FS) ([]string, error) {
	if _, err := db.Exec(schemaMigrationsDDL); err != nil {
		return nil, fmt.Errorf("failed to create schema_migrations: %w", err)
	}

	applied, err := appliedVersions(db)
	if err != nil {
		return nil, err
	}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations: %w", err)
	}

	var ran []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || applied[name] {
			continue
		}

		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", name, err)
		}

		if err := applyMigration(db, name, string(content)); err != nil {
			return nil, err
		}

		ran = append(ran, name)
	}

	return ran, nil
}

func appliedVersions(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}
	return applied, rows.Err()
}

// applyMigration runs a single migration and records it, atomically.
func applyMigration(db *sql.DB, name, content string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for %s: %w", name, err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(content); err != nil {
		return fmt.Errorf("failed to execute %s: %w", name, err)
	}

	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", name); err != nil {
		return fmt.Errorf("failed to record %s: %w", name, err)
	}

	return tx.Commit()
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
