package repo

import "fmt"

type BookInput struct {
	Title            string
	BookType         string
	GenreID          int
	DatePublished    string
	OriginalLanguage string
	URL              string

	Authors []AuthorInput
	Series  *SeriesInput
	TagIDs  []int
}

// AddBook creates a work and, when firstCopy is non-nil, its first copy.
func (r *Repo) AddBook(input *BookInput, firstCopy *CopyInput) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var seriesID *int
	var seriesPosition *float64
	if input.Series != nil {
		id := input.Series.ID
		if id == 0 {
			id, err = createSeries(tx, input.Series.Name)
			if err != nil {
				return fmt.Errorf("failed to create series: %w", err)
			}
		}
		seriesID = &id
		seriesPosition = &input.Series.Position
	}

	result, err := tx.Exec(
		`INSERT INTO book (title, date_published, original_language, type, genre_id, series_id, series_position, url)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Title,
		input.DatePublished,
		input.OriginalLanguage,
		input.BookType,
		input.GenreID,
		seriesID,
		seriesPosition,
		nullable(input.URL),
	)
	if err != nil {
		return fmt.Errorf("failed to create book: %w", err)
	}
	bookID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get book id: %w", err)
	}

	for _, author := range input.Authors {
		authorID := author.ID
		if authorID == 0 {
			authorID, err = createAuthor(tx, author.Name, author.SortName)
			if err != nil {
				return fmt.Errorf("failed to create author: %w", err)
			}
		}
		if err := linkBookAuthor(tx, int(bookID), authorID, author.Role); err != nil {
			return fmt.Errorf("failed to link author: %w", err)
		}
	}

	for _, tagID := range input.TagIDs {
		if err := linkBookTag(tx, int(bookID), tagID); err != nil {
			return fmt.Errorf("failed to link tag: %w", err)
		}
	}

	if firstCopy != nil {
		if _, err := createCopy(tx, int(bookID), firstCopy); err != nil {
			return fmt.Errorf("failed to create copy: %w", err)
		}
	}

	return tx.Commit()
}

// BookEdit is a work's editable fields, without copies, authors or tags.
type BookEdit struct {
	Title            string
	BookType         string
	GenreID          int
	DatePublished    string
	OriginalLanguage string
	URL              string
}

func (r *Repo) GetBook(bookID int) (*BookEdit, error) {
	var b BookEdit
	var url *string
	err := r.db.QueryRow(
		"SELECT title, type, genre_id, date_published, original_language, url FROM book WHERE id = ?",
		bookID,
	).Scan(&b.Title, &b.BookType, &b.GenreID, &b.DatePublished, &b.OriginalLanguage, &url)
	if err != nil {
		return nil, err
	}
	b.URL = deref(url)
	return &b, nil
}

func (r *Repo) UpdateBook(bookID int, in *BookEdit) error {
	_, err := r.db.Exec(
		`UPDATE book SET title = ?, type = ?, genre_id = ?, date_published = ?, original_language = ?, url = ?
         WHERE id = ?`,
		in.Title, in.BookType, in.GenreID, in.DatePublished, in.OriginalLanguage, nullable(in.URL), bookID,
	)
	return err
}

type BookSummary struct {
	ID     int
	Title  string
	Author string
}

func (r *Repo) ListBooks(withCopies bool) ([]BookSummary, error) {
	query := `SELECT b.id, b.title, a.name
        FROM book b
        LEFT JOIN book_author ba ON ba.book_id = b.id AND ba.role = 'author'
        LEFT JOIN author a ON a.id = ba.author_id`

	if withCopies {
		query += `
        WHERE EXISTS (SELECT 1 FROM book_copy bc WHERE bc.book_id = b.id)`
	}

	query += `
        ORDER BY b.title`

	rows, err := r.db.Query(query)
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
