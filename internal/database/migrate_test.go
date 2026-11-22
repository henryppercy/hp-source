package database

import (
	"database/sql"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"

	_ "modernc.org/sqlite"
)

// openTempDB opens an isolated, file-backed sqlite db in a temp dir.
func openTempDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open temp db: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func countRows(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return n
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	return countRows(t, db,
		"SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", name) > 0
}

// mockInit and mockNote are mock migrations used across the tests.
var mockInit = &fstest.MapFile{Data: []byte(
	`CREATE TABLE item (id INTEGER PRIMARY KEY, name TEXT UNIQUE);
	 INSERT INTO item (name) VALUES ('a'), ('b'), ('c');`)}

var mockNote = &fstest.MapFile{Data: []byte(
	`CREATE TABLE note (id INTEGER PRIMARY KEY);`)}

func TestMigrateAppliesAndRecords(t *testing.T) {
	db := openTempDB(t)

	fsys := fstest.MapFS{"001_init.sql": mockInit}
	applied, err := migrateFS(db, fsys)
	if err != nil {
		t.Fatalf("migrateFS: %v", err)
	}

	if !slices.Equal(applied, []string{"001_init.sql"}) {
		t.Errorf("applied = %v, want [001_init.sql]", applied)
	}
	if !tableExists(t, db, "item") {
		t.Error("expected item table to exist after migrate")
	}
	if n := countRows(t, db,
		"SELECT count(*) FROM schema_migrations WHERE version=?", "001_init.sql"); n != 1 {
		t.Errorf("expected 001 recorded once, got %d", n)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	db := openTempDB(t)
	fsys := fstest.MapFS{"001_init.sql": mockInit}

	if _, err := migrateFS(db, fsys); err != nil {
		t.Fatalf("first migrateFS: %v", err)
	}
	items := countRows(t, db, "SELECT count(*) FROM item")

	applied, err := migrateFS(db, fsys)
	if err != nil {
		t.Fatalf("second migrateFS: %v", err)
	}
	if len(applied) != 0 {
		t.Errorf("second run applied %v, want none", applied)
	}
	if n := countRows(t, db, "SELECT count(*) FROM schema_migrations"); n != 1 {
		t.Errorf("expected 1 migration recorded, got %d", n)
	}
	if n := countRows(t, db, "SELECT count(*) FROM item"); n != items {
		t.Errorf("seed duplicated: was %d, now %d", items, n)
	}
}

func TestMigrateIncremental(t *testing.T) {
	db := openTempDB(t)

	fsys := fstest.MapFS{
		"001_init.sql": mockInit,
		"002_note.sql": mockNote,
	}
	applied, err := migrateFS(db, fsys)
	if err != nil {
		t.Fatalf("migrateFS 001+002: %v", err)
	}
	if !slices.Equal(applied, []string{"001_init.sql", "002_note.sql"}) {
		t.Fatalf("applied = %v, want [001_init.sql 002_note.sql]", applied)
	}

	fsys["003_extra.sql"] = &fstest.MapFile{Data: []byte("CREATE TABLE extra (id INTEGER PRIMARY KEY);")}
	applied, err = migrateFS(db, fsys)
	if err != nil {
		t.Fatalf("migrateFS 003: %v", err)
	}
	if !slices.Equal(applied, []string{"003_extra.sql"}) {
		t.Errorf("applied = %v, want [003_extra.sql]", applied)
	}
	if !tableExists(t, db, "extra") {
		t.Error("expected extra table to exist")
	}
}

func TestMigrateAdoptPreexistingSchema(t *testing.T) {
	fsys := fstest.MapFS{
		"001_init.sql": mockInit,
		"002_note.sql": mockNote,
	}

	raw := openTempDB(t)
	if _, err := raw.Exec(string(mockInit.Data)); err != nil {
		t.Fatalf("seed preexisting schema: %v", err)
	}
	if _, err := migrateFS(raw, fsys); err == nil {
		t.Fatal("expected migrateFS to fail on un-baselined preexisting schema")
	}

	db := openTempDB(t)
	if _, err := db.Exec(string(mockInit.Data)); err != nil {
		t.Fatalf("seed preexisting schema: %v", err)
	}
	items := countRows(t, db, "SELECT count(*) FROM item")

	baseline := schemaMigration + `;
	INSERT OR IGNORE INTO schema_migrations (version) VALUES ('001_init.sql');`
	if _, err := db.Exec(baseline); err != nil {
		t.Fatalf("baseline statement: %v", err)
	}

	applied, err := migrateFS(db, fsys)
	if err != nil {
		t.Fatalf("migrateFS after baseline: %v", err)
	}
	if !slices.Equal(applied, []string{"002_note.sql"}) {
		t.Errorf("applied = %v, want [002_note.sql]", applied)
	}
	if n := countRows(t, db, "SELECT count(*) FROM item"); n != items {
		t.Errorf("existing data changed: was %d, now %d", items, n)
	}
}
