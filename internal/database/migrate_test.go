package database

import (
	"database/sql"
	"path/filepath"
	"testing"
	"testing/fstest"

	_ "modernc.org/sqlite"
)

// Opens an isolated, file-backed sqlite db in a temp dir.
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

// mock migrations used across the tests
var mockInit = &fstest.MapFile{Data: []byte(
	`CREATE TABLE item (id INTEGER PRIMARY KEY, name TEXT UNIQUE);
	 INSERT INTO item (name) VALUES ('a'), ('b'), ('c');`)}

var mockNote = &fstest.MapFile{Data: []byte(
	`CREATE TABLE note (id INTEGER PRIMARY KEY);`)}

func TestMigrateAppliesAndRecords(t *testing.T) {
	db := openTempDB(t)

	fsys := fstest.MapFS{"001_init.sql": mockInit}
	if err := migrateFS(db, fsys); err != nil {
		t.Fatalf("migrateFS: %v", err)
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

	if err := migrateFS(db, fsys); err != nil {
		t.Fatalf("first migrateFS: %v", err)
	}
	items := countRows(t, db, "SELECT count(*) FROM item")

	if err := migrateFS(db, fsys); err != nil {
		t.Fatalf("second migrateFS: %v", err)
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
	if err := migrateFS(db, fsys); err != nil {
		t.Fatalf("migrateFS 001+002: %v", err)
	}
	if !tableExists(t, db, "item") || !tableExists(t, db, "note") {
		t.Fatal("expected item and note tables to exist")
	}
	if n := countRows(t, db, "SELECT count(*) FROM schema_migrations"); n != 2 {
		t.Fatalf("expected 2 recorded, got %d", n)
	}

	// Add a third migration; only it should apply. If 001/002 re-ran, their
	// CREATE TABLE statements would fail on the already-existing tables.
	fsys["003_extra.sql"] = &fstest.MapFile{Data: []byte("CREATE TABLE extra (id INTEGER PRIMARY KEY);")}
	if err := migrateFS(db, fsys); err != nil {
		t.Fatalf("migrateFS 003: %v", err)
	}
	if !tableExists(t, db, "extra") {
		t.Error("expected extra table to exist")
	}
	if n := countRows(t, db, "SELECT count(*) FROM schema_migrations"); n != 3 {
		t.Errorf("expected 3 recorded, got %d", n)
	}
}

// Models adopting a db that already has a migration's schema but no schema_migrations
func TestMigrateAdoptPreexistingSchema(t *testing.T) {
	fsys := fstest.MapFS{
		"001_init.sql": mockInit,
		"002_note.sql": mockNote,
	}

	// A db where 001's schema already exists but nothing records that it ran:
	// migrateFS must fail trying to re-run 001 (table already exists).
	raw := openTempDB(t)
	if _, err := raw.Exec(string(mockInit.Data)); err != nil {
		t.Fatalf("seed preexisting schema: %v", err)
	}
	if err := migrateFS(raw, fsys); err == nil {
		t.Fatal("expected migrateFS to fail on un-baselined preexisting schema")
	}

	// Same starting point, but first apply the documented baseline: create
	// schema_migrations and mark 001 as already applied. migrateFS should then
	// skip 001, apply 002, and leave existing data intact.
	db := openTempDB(t)
	if _, err := db.Exec(string(mockInit.Data)); err != nil {
		t.Fatalf("seed preexisting schema: %v", err)
	}
	items := countRows(t, db, "SELECT count(*) FROM item")

	baseline := schemaMigrationsDDL + `;
	INSERT OR IGNORE INTO schema_migrations (version) VALUES ('001_init.sql');`
	if _, err := db.Exec(baseline); err != nil {
		t.Fatalf("baseline statement: %v", err)
	}

	if err := migrateFS(db, fsys); err != nil {
		t.Fatalf("migrateFS after baseline: %v", err)
	}
	if !tableExists(t, db, "note") {
		t.Error("expected 002 to have applied (note table)")
	}
	if n := countRows(t, db, "SELECT count(*) FROM item"); n != items {
		t.Errorf("existing data changed: was %d, now %d", items, n)
	}
}
