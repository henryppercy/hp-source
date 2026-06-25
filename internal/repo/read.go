package repo

import "fmt"

type ReadInput struct {
	BookID       int
	CopyID       int
	Status       string
	Rating       int
	DateStarted  string
	DateFinished string
}

type ActiveRead struct {
	ReadID    int
	BookTitle string
	Author    string
	Format    string
}

// ReadEntry is a read with the book detail the site's reading page needs.
type ReadEntry struct {
	Title        string
	Author       string
	CoverImage   string
	Status       string
	Rating       int
	DateFinished string
}

type StartReadInput struct {
	BookID      int
	CopyID      int
	DateStarted string
}

type FinishReadInput struct {
	ReadID       int
	Status       string
	Rating       int
	DateFinished string
}

func (r *Repo) AddRead(input *ReadInput) error {
	_, err := r.db.Exec(
		`INSERT INTO read (book_id, copy_id, status, rating, date_started, date_finished)
         VALUES (?, ?, ?, ?, ?, ?)`,
		input.BookID,
		input.CopyID,
		input.Status,
		nullableInt(input.Rating),
		nullable(input.DateStarted),
		nullable(input.DateFinished),
	)
	if err != nil {
		return fmt.Errorf("failed to add read: %w", err)
	}
	return nil
}

func (r *Repo) ListBooksAvailableToRead() ([]BookSummary, error) {
	rows, err := r.db.Query(
		`SELECT b.id, b.title, a.name
         FROM book b
         LEFT JOIN book_author ba ON ba.book_id = b.id AND ba.role = 'author'
         LEFT JOIN author a ON a.id = ba.author_id
         WHERE EXISTS (SELECT 1 FROM book_copy bc WHERE bc.book_id = b.id)
         AND NOT EXISTS (SELECT 1 FROM read r WHERE r.book_id = b.id AND r.status = 'reading')
         ORDER BY b.title`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list books available to read: %w", err)
	}
	defer rows.Close()

	var books []BookSummary
	for rows.Next() {
		var b BookSummary
		var author *string
		if err := rows.Scan(&b.ID, &b.Title, &author); err != nil {
			return nil, fmt.Errorf("failed to scan book: %w", err)
		}
		if author != nil {
			b.Author = *author
		}
		books = append(books, b)
	}
	return books, nil
}

func (r *Repo) ListActiveReads() ([]ActiveRead, error) {
	rows, err := r.db.Query(
		`SELECT rd.id, b.title, a.name, bc.format
         FROM read rd
         JOIN book b ON b.id = rd.book_id
         LEFT JOIN book_author ba ON ba.book_id = b.id AND ba.role = 'author'
         LEFT JOIN author a ON a.id = ba.author_id
         LEFT JOIN book_copy bc ON bc.id = rd.copy_id
         WHERE rd.status = 'reading'
         ORDER BY rd.date_started DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list active reads: %w", err)
	}
	defer rows.Close()

	var reads []ActiveRead
	for rows.Next() {
		var r ActiveRead
		var author, format *string
		if err := rows.Scan(&r.ReadID, &r.BookTitle, &author, &format); err != nil {
			return nil, fmt.Errorf("failed to scan read: %w", err)
		}
		if author != nil {
			r.Author = *author
		}
		if format != nil {
			r.Format = *format
		}
		reads = append(reads, r)
	}
	return reads, nil
}

// ListReads returns all reads with book detail for the site, newest finished
// first. The site groups them by status.
func (r *Repo) ListReads() ([]ReadEntry, error) {
	rows, err := r.db.Query(
		`SELECT b.title, a.name, bc.cover_image, rd.status, rd.rating, rd.date_finished
         FROM read rd
         JOIN book b ON b.id = rd.book_id
         LEFT JOIN book_author ba ON ba.book_id = b.id AND ba.role = 'author'
         LEFT JOIN author a ON a.id = ba.author_id
         LEFT JOIN book_copy bc ON bc.id = rd.copy_id
         ORDER BY rd.date_finished DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list reads: %w", err)
	}
	defer rows.Close()

	var entries []ReadEntry
	for rows.Next() {
		var e ReadEntry
		var author, coverImage, dateFinished *string
		var rating *int
		if err := rows.Scan(&e.Title, &author, &coverImage, &e.Status, &rating, &dateFinished); err != nil {
			return nil, fmt.Errorf("failed to scan read: %w", err)
		}
		if author != nil {
			e.Author = *author
		}
		if coverImage != nil {
			e.CoverImage = *coverImage
		}
		if rating != nil {
			e.Rating = *rating
		}
		if dateFinished != nil {
			e.DateFinished = *dateFinished
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (r *Repo) StartRead(bookID, copyID int, dateStarted string) error {
	_, err := r.db.Exec(
		`INSERT INTO read (book_id, copy_id, status, date_started)
         VALUES (?, ?, 'reading', ?)`,
		bookID, copyID, dateStarted,
	)
	if err != nil {
		return fmt.Errorf("failed to start read: %w", err)
	}
	return nil
}

func (r *Repo) FinishRead(readID int, status string, rating int, dateFinished string) error {
	_, err := r.db.Exec(
		`UPDATE read SET status = ?, rating = ?, date_finished = ?
         WHERE id = ?`,
		status, nullableInt(rating), nullable(dateFinished), readID,
	)
	if err != nil {
		return fmt.Errorf("failed to finish read: %w", err)
	}
	return nil
}
