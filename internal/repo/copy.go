package repo

import "fmt"

type CopyInput struct {
	Title        string
	Headline     string
	ShelfStatus  string
	Format       string
	PageCount    int
	Language     string
	ISBN         string
	CoverImage   string
	Source       string
	DateAcquired string
	SecondHand   bool
	Translators  []TranslatorInput
}

func createCopy(tx TX, bookID int, in *CopyInput) (int, error) {
	result, err := tx.Exec(
		`INSERT INTO book_copy (book_id, title, headline, shelf_status, format, page_count, language, isbn, cover_image, source, date_acquired, second_hand)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		bookID,
		in.Title,
		nullable(in.Headline),
		nullable(in.ShelfStatus),
		in.Format,
		nullableInt(in.PageCount),
		nullable(in.Language),
		nullable(in.ISBN),
		nullable(in.CoverImage),
		nullable(in.Source),
		nullable(in.DateAcquired),
		boolInt(in.SecondHand),
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	if err := linkTranslators(tx, int(id), in.Translators); err != nil {
		return 0, err
	}
	return int(id), nil
}

// AddCopy attaches a new copy to an existing book.
func (r *Repo) AddCopy(bookID int, in *CopyInput) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := createCopy(tx, bookID, in); err != nil {
		return fmt.Errorf("failed to create copy: %w", err)
	}
	return tx.Commit()
}

func (r *Repo) GetCopy(copyID int) (*CopyInput, error) {
	var in CopyInput
	var pageCount, secondHand *int
	var headline, shelfStatus, language, isbn, coverImage, source, dateAcquired *string
	err := r.db.QueryRow(
		`SELECT title, headline, shelf_status, format, page_count, language, isbn, cover_image, source, date_acquired, second_hand
         FROM book_copy WHERE id = ?`,
		copyID,
	).Scan(&in.Title, &headline, &shelfStatus, &in.Format, &pageCount, &language, &isbn, &coverImage, &source, &dateAcquired, &secondHand)
	if err != nil {
		return nil, err
	}
	if pageCount != nil {
		in.PageCount = *pageCount
	}
	in.Headline = deref(headline)
	in.ShelfStatus = deref(shelfStatus)
	in.Language = deref(language)
	in.ISBN = deref(isbn)
	in.CoverImage = deref(coverImage)
	in.Source = deref(source)
	in.DateAcquired = deref(dateAcquired)
	in.SecondHand = secondHand != nil && *secondHand == 1

	translators, err := loadCopyTranslators(r.db, copyID)
	if err != nil {
		return nil, err
	}
	in.Translators = translators
	return &in, nil
}

func (r *Repo) UpdateCopy(copyID int, in *CopyInput) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE book_copy
         SET title = ?, headline = ?, shelf_status = ?, format = ?, page_count = ?, language = ?, isbn = ?, cover_image = ?, source = ?, date_acquired = ?, second_hand = ?
         WHERE id = ?`,
		in.Title,
		nullable(in.Headline),
		nullable(in.ShelfStatus),
		in.Format,
		nullableInt(in.PageCount),
		nullable(in.Language),
		nullable(in.ISBN),
		nullable(in.CoverImage),
		nullable(in.Source),
		nullable(in.DateAcquired),
		boolInt(in.SecondHand),
		copyID,
	)
	if err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM copy_translator WHERE copy_id = ?", copyID); err != nil {
		return err
	}
	if err := linkTranslators(tx, copyID, in.Translators); err != nil {
		return err
	}
	return tx.Commit()
}

type Copy struct {
	ID       int
	Title    string
	Format   string
	Language string
}

func (r *Repo) ListCopies(bookID int) ([]Copy, error) {
	rows, err := r.db.Query(
		"SELECT id, title, format, language FROM book_copy WHERE book_id = ? ORDER BY id",
		bookID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var copies []Copy
	for rows.Next() {
		var c Copy
		if err := rows.Scan(&c.ID, &c.Title, &c.Format, &c.Language); err != nil {
			return nil, err
		}
		copies = append(copies, c)
	}
	return copies, nil
}
