package repo

import (
	"database/sql"
	"errors"
	"fmt"
)

// BookPage is a book's full detail for its page, the work with its oldest
// copy merged in, plus every read of it.
type BookPage struct {
	ID               int
	Title            string
	OriginalTitle    string // in the original language, empty when that is english
	Authors          []BookCredit
	Translators      []string
	BookType         string
	Genre            string
	DatePublished    string
	OriginalLanguage string
	Series           string
	SeriesPosition   *float64
	Tags             []string

	Format       string
	Language     string
	PageCount    int
	CoverImage   string
	Source       string
	SecondHand   bool
	DateAcquired string
	ShelfStatus  string

	Reads []BookRead
}

// BookCredit is one author's name and role on a book.
type BookCredit struct {
	Name string
	Role string
}

// BookRead is one read of a book with the edition detail of the copy it used.
type BookRead struct {
	Status       string
	Rating       int
	DateStarted  string
	DateFinished string
	Title        string // the read's copy title, so another edition shows on its line
	Language     string
	Format       string
	PageCount    int
	CurrentPage  int
}

// ListBookPageIDs returns the id of every book with a copy or a read, in
// title order. These are the books that get pages.
func (r *Repo) ListBookPageIDs() ([]int, error) {
	rows, err := r.db.Query(
		`SELECT b.id
         FROM book b
         WHERE EXISTS (SELECT 1 FROM book_copy bc WHERE bc.book_id = b.id)
            OR EXISTS (SELECT 1 FROM read rd WHERE rd.book_id = b.id)
         ORDER BY b.title, b.id`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list book page ids: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan book page id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetBookPage returns the book's full page detail, erroring when none exists.
func (r *Repo) GetBookPage(bookID int) (*BookPage, error) {
	p := &BookPage{ID: bookID}
	var series, originalTitle *string
	var seriesPos *float64
	err := r.db.QueryRow(
		`SELECT b.title, b.original_title, b.type, g.name, b.date_published, b.original_language, s.name, b.series_position
         FROM book b
         JOIN genre g ON g.id = b.genre_id
         LEFT JOIN series s ON s.id = b.series_id
         WHERE b.id = ?`,
		bookID,
	).Scan(&p.Title, &originalTitle, &p.BookType, &p.Genre, &p.DatePublished, &p.OriginalLanguage, &series, &seriesPos)
	if err != nil {
		return nil, fmt.Errorf("failed to get book page: %w", err)
	}
	p.OriginalTitle = deref(originalTitle)
	p.Series = deref(series)
	p.SeriesPosition = seriesPos

	if err := r.loadBookCredits(p); err != nil {
		return nil, err
	}
	if err := r.loadBookTags(p); err != nil {
		return nil, err
	}
	if err := r.loadDisplayCopy(p); err != nil {
		return nil, err
	}
	return p, r.loadBookReads(p)
}

func (r *Repo) loadBookCredits(p *BookPage) error {
	rows, err := r.db.Query(
		`SELECT a.name, ba.role
         FROM book_author ba
         JOIN author a ON a.id = ba.author_id
         WHERE ba.book_id = ?
         ORDER BY a.sort_name, ba.role`,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to load book authors: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var c BookCredit
		if err := rows.Scan(&c.Name, &c.Role); err != nil {
			return fmt.Errorf("failed to scan book author: %w", err)
		}
		p.Authors = append(p.Authors, c)
	}
	return rows.Err()
}

func (r *Repo) loadBookTags(p *BookPage) error {
	rows, err := r.db.Query(
		`SELECT t.name
         FROM book_tag bt
         JOIN tag t ON t.id = bt.tag_id
         WHERE bt.book_id = ?
         ORDER BY t.name`,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to load book tags: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("failed to scan book tag: %w", err)
		}
		p.Tags = append(p.Tags, name)
	}
	return rows.Err()
}

// loadDisplayCopy merges the oldest copy's fields and translators into the
// page. A book with no copy keeps only the work's own fields.
func (r *Repo) loadDisplayCopy(p *BookPage) error {
	var format, language, coverImage, source, dateAcquired, shelfStatus *string
	var pageCount, secondHand *int
	var copyID int
	err := r.db.QueryRow(
		`SELECT id, format, language, page_count, cover_image, source, second_hand, date_acquired, shelf_status
         FROM book_copy
         WHERE book_id = ?
         ORDER BY id
         LIMIT 1`,
		p.ID,
	).Scan(&copyID, &format, &language, &pageCount, &coverImage, &source, &secondHand, &dateAcquired, &shelfStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to load display copy: %w", err)
	}

	p.Format = deref(format)
	p.Language = deref(language)
	p.CoverImage = deref(coverImage)
	p.Source = deref(source)
	p.DateAcquired = deref(dateAcquired)
	p.ShelfStatus = deref(shelfStatus)
	p.SecondHand = secondHand != nil && *secondHand == 1
	if pageCount != nil {
		p.PageCount = *pageCount
	}

	rows, err := r.db.Query(
		`SELECT tr.name
         FROM copy_translator ct
         JOIN translator tr ON tr.id = ct.translator_id
         WHERE ct.copy_id = ?
         ORDER BY tr.sort_name`,
		copyID,
	)
	if err != nil {
		return fmt.Errorf("failed to load copy translators: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("failed to scan copy translator: %w", err)
		}
		p.Translators = append(p.Translators, name)
	}
	return rows.Err()
}

// loadBookReads fills every read, the one in progress first, each carrying
// its own copy's detail.
func (r *Repo) loadBookReads(p *BookPage) error {
	rows, err := r.db.Query(
		`SELECT rd.status, rd.rating, rd.date_started, rd.date_finished,
                bc.title, bc.language, bc.format, bc.page_count,
                (SELECT rl.page FROM read_log rl
                 WHERE rl.read_id = rd.id AND rl.page IS NOT NULL
                 ORDER BY rl.id DESC LIMIT 1)
         FROM read rd
         LEFT JOIN book_copy bc ON bc.id = rd.copy_id
         WHERE rd.book_id = ?
         ORDER BY CASE rd.status WHEN 'reading' THEN 0 ELSE 1 END,
                  rd.date_finished DESC, rd.date_started DESC, rd.id DESC`,
		p.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to load book reads: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rd BookRead
		var rating, pageCount, currentPage *int
		var title, language, format, dateStarted, dateFinished *string
		if err := rows.Scan(
			&rd.Status, &rating, &dateStarted, &dateFinished,
			&title, &language, &format, &pageCount, &currentPage,
		); err != nil {
			return fmt.Errorf("failed to scan book read: %w", err)
		}
		rd.DateStarted = deref(dateStarted)
		rd.DateFinished = deref(dateFinished)
		rd.Title = deref(title)
		rd.Language = deref(language)
		rd.Format = deref(format)
		if rating != nil {
			rd.Rating = *rating
		}
		if pageCount != nil {
			rd.PageCount = *pageCount
		}
		if currentPage != nil {
			rd.CurrentPage = *currentPage
		}
		p.Reads = append(p.Reads, rd)
	}
	return rows.Err()
}
