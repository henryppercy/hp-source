package repo

import "database/sql"

// Repo runs the app's queries against the SQLite database.
type Repo struct {
	db *sql.DB
}

// TX is the query surface shared by *sql.DB and *sql.Tx, so helpers can run
// inside or outside a transaction.
type TX interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
}

// New returns a Repo backed by db.
func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}
