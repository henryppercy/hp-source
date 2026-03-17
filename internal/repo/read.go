package repo

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
	return err
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
		return nil, err
	}
	defer rows.Close()

	var books []BookSummary
	for rows.Next() {
		var b BookSummary
		var author *string
		if err := rows.Scan(&b.ID, &b.Title, &author); err != nil {
			return nil, err
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
		return nil, err
	}
	defer rows.Close()

	var reads []ActiveRead
	for rows.Next() {
		var r ActiveRead
		var author, format *string
		if err := rows.Scan(&r.ReadID, &r.BookTitle, &author, &format); err != nil {
			return nil, err
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

func (r *Repo) StartRead(bookID, copyID int, dateStarted string) error {
	_, err := r.db.Exec(
		`INSERT INTO read (book_id, copy_id, status, date_started)
         VALUES (?, ?, 'reading', ?)`,
		bookID, copyID, dateStarted,
	)
	return err
}

func (r *Repo) FinishRead(readID int, status string, rating int, dateFinished string) error {
	_, err := r.db.Exec(
		`UPDATE read SET status = ?, rating = ?, date_finished = ?
         WHERE id = ?`,
		status, nullableInt(rating), nullable(dateFinished), readID,
	)
	return err
}
