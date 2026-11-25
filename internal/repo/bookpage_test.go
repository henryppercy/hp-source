package repo

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/henryppercy/hp-source/internal/database"
)

// newTestRepo opens a temp sqlite file, migrates it, and wraps it in a Repo.
func newTestRepo(t *testing.T) *Repo {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open temp db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return New(db)
}

// seedBookPage creates a translated book with an english title, a russian
// original, an older english copy carrying a translator and a newer spanish
// one, reads against both copies, an active read logged to page 100, a
// shelved second book and a bare third. It returns the translated book id.
func seedBookPage(t *testing.T, r *Repo) int {
	t.Helper()
	db := r.db

	var genreID int
	if err := db.QueryRow("SELECT id FROM genre WHERE name = 'classics'").Scan(&genreID); err != nil {
		t.Fatalf("genre: %v", err)
	}
	res, err := db.Exec(
		`INSERT INTO book (title, original_title, date_published, original_language, type, genre_id)
         VALUES ('Crime and Punishment', 'Преступление и наказание', '1866-01-01', 'russian', 'fiction', ?)`, genreID)
	if err != nil {
		t.Fatalf("book: %v", err)
	}
	bookID, _ := res.LastInsertId()

	if _, err := db.Exec("INSERT INTO translator (name, sort_name) VALUES ('Richard Pevear', 'Pevear, Richard')"); err != nil {
		t.Fatalf("translator: %v", err)
	}
	var translatorID int
	if err := db.QueryRow("SELECT id FROM translator").Scan(&translatorID); err != nil {
		t.Fatalf("translator id: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO book_copy (book_id, title, format, page_count, language, source, date_acquired)
         VALUES (?, 'Crime and Punishment', 'paperback', 564, 'english', 'bought', '2020-01-01')`, bookID,
	); err != nil {
		t.Fatalf("english copy: %v", err)
	}
	var engCopy int
	if err := db.QueryRow("SELECT id FROM book_copy WHERE language = 'english'").Scan(&engCopy); err != nil {
		t.Fatalf("english copy id: %v", err)
	}
	if _, err := db.Exec("INSERT INTO copy_translator (copy_id, translator_id) VALUES (?, ?)", engCopy, translatorID); err != nil {
		t.Fatalf("link translator: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO book_copy (book_id, title, format, page_count, language, source)
         VALUES (?, 'Crimen y castigo', 'paperback', 672, 'spanish', 'bought')`, bookID,
	); err != nil {
		t.Fatalf("spanish copy: %v", err)
	}
	var espCopy int
	if err := db.QueryRow("SELECT id FROM book_copy WHERE language = 'spanish'").Scan(&espCopy); err != nil {
		t.Fatalf("spanish copy id: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO read (book_id, copy_id, status, rating, date_started, date_finished)
         VALUES (?, ?, 'finished', 9, '2021-02-01', '2021-02-20')`, bookID, engCopy,
	); err != nil {
		t.Fatalf("english read: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO read (book_id, copy_id, status, rating, date_started, date_finished)
         VALUES (?, ?, 'finished', 7, '2025-06-01', '2025-06-30')`, bookID, espCopy,
	); err != nil {
		t.Fatalf("spanish read: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO read (book_id, copy_id, status, date_started) VALUES (?, ?, 'reading', '2026-01-01')`,
		bookID, engCopy,
	); err != nil {
		t.Fatalf("active read: %v", err)
	}
	var activeRead int
	if err := db.QueryRow("SELECT id FROM read WHERE status = 'reading'").Scan(&activeRead); err != nil {
		t.Fatalf("active read id: %v", err)
	}
	if _, err := db.Exec("INSERT INTO read_log (read_id, page) VALUES (?, 100)", activeRead); err != nil {
		t.Fatalf("read log: %v", err)
	}

	if _, err := db.Exec(
		"INSERT INTO book_tag (book_id, tag_id) VALUES (?, (SELECT id FROM tag WHERE name='noir'))", bookID,
	); err != nil {
		t.Fatalf("link tag: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO book (title, date_published, original_language, type, genre_id)
         VALUES ('Alone', '2014-01-01', 'english', 'non-fiction', ?)`, genreID,
	); err != nil {
		t.Fatalf("shelved book: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO book_copy (book_id, title, format, language, shelf_status)
         VALUES ((SELECT max(id) FROM book), 'Alone', 'hardback', 'english', 'shelf')`,
	); err != nil {
		t.Fatalf("shelved copy: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO book (title, date_published, original_language, type, genre_id)
         VALUES ('Nothing', '2014-01-01', 'english', 'fiction', ?)`, genreID,
	); err != nil {
		t.Fatalf("bare book: %v", err)
	}

	return int(bookID)
}

func TestListBookPageIDs(t *testing.T) {
	r := newTestRepo(t)
	bookID := seedBookPage(t, r)

	ids, err := r.ListBookPageIDs()
	if err != nil {
		t.Fatalf("ListBookPageIDs: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("got %d ids %v, want 2, the translated book and the shelved one", len(ids), ids)
	}
	for _, id := range ids {
		if id == bookID {
			return
		}
	}
	t.Errorf("translated book id %d missing from %v", bookID, ids)
}

func TestGetBookPageTitles(t *testing.T) {
	r := newTestRepo(t)
	bookID := seedBookPage(t, r)

	p, err := r.GetBookPage(bookID)
	if err != nil {
		t.Fatalf("GetBookPage: %v", err)
	}
	if p.Title != "Crime and Punishment" {
		t.Errorf("Title = %q, want the book row's english title", p.Title)
	}
	if p.OriginalTitle != "Преступление и наказание" {
		t.Errorf("OriginalTitle = %q, want the russian one", p.OriginalTitle)
	}
	if p.Format != "paperback" || p.Source != "bought" || p.PageCount != 564 {
		t.Errorf("copy merge wrong, format %q source %q pages %d, want the oldest copy's fields", p.Format, p.Source, p.PageCount)
	}
	if len(p.Translators) != 1 || p.Translators[0] != "Richard Pevear" {
		t.Errorf("Translators = %v, want Pevear from the oldest copy", p.Translators)
	}
	if len(p.Tags) != 1 || p.Tags[0] != "noir" {
		t.Errorf("Tags = %v, want the linked noir tag", p.Tags)
	}
}

func TestGetBookPageReads(t *testing.T) {
	r := newTestRepo(t)
	bookID := seedBookPage(t, r)

	p, err := r.GetBookPage(bookID)
	if err != nil {
		t.Fatalf("GetBookPage: %v", err)
	}
	if len(p.Reads) != 3 {
		t.Fatalf("got %d reads, want 3", len(p.Reads))
	}

	active := p.Reads[0]
	if active.Status != "reading" {
		t.Fatalf("first read = %q, want the active one, it sorts first", active.Status)
	}
	if active.CurrentPage != 100 {
		t.Errorf("active read CurrentPage = %d, want 100 from its log", active.CurrentPage)
	}

	spanish, english := p.Reads[1], p.Reads[2]
	if spanish.Status != "finished" || spanish.Title != "Crimen y castigo" || spanish.PageCount != 672 {
		t.Errorf("newest finished read wrong, %q %q %d pages, want the 2025 spanish one", spanish.Status, spanish.Title, spanish.PageCount)
	}
	if english.Title != "Crime and Punishment" || english.Rating != 9 || english.CurrentPage != 0 {
		t.Errorf("older finished read wrong, %q rating %d page %d, want the english copy with 9 and no logged page", english.Title, english.Rating, english.CurrentPage)
	}
}
