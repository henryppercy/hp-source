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
	Genre        string
	BookType     string
	Format       string
	Source       string
	SecondHand   bool
	Status       string
	Rating       int
	PageCount    int
	CurrentPage  int
	DateStarted  string
	DateFinished string
}

// ShelfEntry is an owned book with no read yet: the antilibrary.
type ShelfEntry struct {
	Title        string
	Author       string
	CoverImage   string
	Genre        string
	Format       string
	PageCount    int
	DateAcquired string
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
		`SELECT rd.id, COALESCE(bc.title, b.title), a.name, bc.format
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
		`SELECT COALESCE(bc.title, b.title), a.name, bc.cover_image, g.name, b.type, bc.format, bc.source, bc.second_hand,
                rd.status, rd.rating, bc.page_count,
                (SELECT rl.page FROM read_log rl
                 WHERE rl.read_id = rd.id AND rl.page IS NOT NULL
                 ORDER BY rl.id DESC LIMIT 1),
                rd.date_started, rd.date_finished
         FROM read rd
         JOIN book b ON b.id = rd.book_id
         JOIN genre g ON g.id = b.genre_id
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
		var author, coverImage, format, source, dateStarted, dateFinished *string
		var rating, pageCount, currentPage, secondHand *int
		if err := rows.Scan(
			&e.Title, &author, &coverImage, &e.Genre, &e.BookType, &format, &source, &secondHand,
			&e.Status, &rating, &pageCount, &currentPage, &dateStarted, &dateFinished,
		); err != nil {
			return nil, fmt.Errorf("failed to scan read: %w", err)
		}
		if author != nil {
			e.Author = *author
		}
		if coverImage != nil {
			e.CoverImage = *coverImage
		}
		if format != nil {
			e.Format = *format
		}
		if source != nil {
			e.Source = *source
		}
		e.SecondHand = secondHand != nil && *secondHand == 1
		if rating != nil {
			e.Rating = *rating
		}
		if pageCount != nil {
			e.PageCount = *pageCount
		}
		if currentPage != nil {
			e.CurrentPage = *currentPage
		}
		if dateStarted != nil {
			e.DateStarted = *dateStarted
		}
		if dateFinished != nil {
			e.DateFinished = *dateFinished
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// ListShelf returns shelved copies not yet read, newest acquisition first. It is
// per copy: a book owned in two languages shows as two entries.
func (r *Repo) ListShelf() ([]ShelfEntry, error) {
	rows, err := r.db.Query(
		`SELECT COALESCE(bc.title, b.title), a.name, bc.cover_image, g.name, bc.format,
                bc.page_count, bc.date_acquired
         FROM book_copy bc
         JOIN book b ON b.id = bc.book_id
         JOIN genre g ON g.id = b.genre_id
         LEFT JOIN book_author ba ON ba.book_id = b.id AND ba.role = 'author'
         LEFT JOIN author a ON a.id = ba.author_id
         WHERE bc.shelf_status = 'shelf'
         AND NOT EXISTS (SELECT 1 FROM read r WHERE r.copy_id = bc.id)
         ORDER BY bc.date_acquired DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list shelf: %w", err)
	}
	defer rows.Close()

	var entries []ShelfEntry
	for rows.Next() {
		var e ShelfEntry
		var author, coverImage, format, dateAcquired *string
		var pageCount *int
		if err := rows.Scan(&e.Title, &author, &coverImage, &e.Genre, &format, &pageCount, &dateAcquired); err != nil {
			return nil, fmt.Errorf("failed to scan shelf book: %w", err)
		}
		if author != nil {
			e.Author = *author
		}
		if coverImage != nil {
			e.CoverImage = *coverImage
		}
		if format != nil {
			e.Format = *format
		}
		if pageCount != nil {
			e.PageCount = *pageCount
		}
		if dateAcquired != nil {
			e.DateAcquired = *dateAcquired
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

type ReadLogInput struct {
	ReadID int
	Page   int
	Note   string
}

func (r *Repo) AddReadLog(in *ReadLogInput) error {
	_, err := r.db.Exec(
		`INSERT INTO read_log (read_id, page, note) VALUES (?, ?, ?)`,
		in.ReadID, nullableInt(in.Page), nullable(in.Note),
	)
	if err != nil {
		return fmt.Errorf("failed to add read log: %w", err)
	}
	return nil
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

// ReadDetail is an existing read with the current values the edit form prefills.
type ReadDetail struct {
	ReadID       int
	BookTitle    string
	Author       string
	Format       string
	Status       string
	Rating       int
	DateStarted  string
	DateFinished string
}

type EditReadInput struct {
	ReadID       int
	Status       string
	Rating       int
	DateStarted  string
	DateFinished string
}

// ListReadsForEdit returns every read with its current values, newest first.
func (r *Repo) ListReadsForEdit() ([]ReadDetail, error) {
	rows, err := r.db.Query(
		`SELECT rd.id, COALESCE(bc.title, b.title), a.name, bc.format,
                rd.status, rd.rating, rd.date_started, rd.date_finished
         FROM read rd
         JOIN book b ON b.id = rd.book_id
         LEFT JOIN book_author ba ON ba.book_id = b.id AND ba.role = 'author'
         LEFT JOIN author a ON a.id = ba.author_id
         LEFT JOIN book_copy bc ON bc.id = rd.copy_id
         ORDER BY rd.date_finished DESC, rd.date_started DESC, rd.id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list reads for edit: %w", err)
	}
	defer rows.Close()

	var reads []ReadDetail
	for rows.Next() {
		var d ReadDetail
		var author, format, dateStarted, dateFinished *string
		var rating *int
		if err := rows.Scan(
			&d.ReadID, &d.BookTitle, &author, &format,
			&d.Status, &rating, &dateStarted, &dateFinished,
		); err != nil {
			return nil, fmt.Errorf("failed to scan read: %w", err)
		}
		if author != nil {
			d.Author = *author
		}
		if format != nil {
			d.Format = *format
		}
		if rating != nil {
			d.Rating = *rating
		}
		if dateStarted != nil {
			d.DateStarted = *dateStarted
		}
		if dateFinished != nil {
			d.DateFinished = *dateFinished
		}
		reads = append(reads, d)
	}
	return reads, rows.Err()
}

func (r *Repo) UpdateRead(in *EditReadInput) error {
	_, err := r.db.Exec(
		`UPDATE read SET status = ?, rating = ?, date_started = ?, date_finished = ?
         WHERE id = ?`,
		in.Status,
		nullableInt(in.Rating),
		nullable(in.DateStarted),
		nullable(in.DateFinished),
		in.ReadID,
	)
	if err != nil {
		return fmt.Errorf("failed to update read: %w", err)
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
