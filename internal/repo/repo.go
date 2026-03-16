package repo

import "database/sql"

type Repo struct {
	db *sql.DB
}

type TX interface {
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
}

func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}
