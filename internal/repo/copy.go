package repo

func createCopy(tx TX, bookID int, format string, pageCount *int, language, isbn, coverImage, source, dateAcquired string, secondHand bool) (int, error) {
	result, err := tx.Exec(
		`INSERT INTO book_copy (book_id, format, page_count, language, isbn, cover_image, source, date_acquired, second_hand)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		bookID,
		format,
		pageCount,
		nullable(language),
		nullable(isbn),
		nullable(coverImage),
		nullable(source),
		nullable(dateAcquired),
		boolInt(secondHand),
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

type CopyInput struct {
	Format       string
	PageCount    int
	Language     string
	ISBN         string
	CoverImage   string
	Source       string
	DateAcquired string
	SecondHand   bool
}

func (r *Repo) GetCopy(copyID int) (*CopyInput, error) {
	var in CopyInput
	var pageCount, secondHand *int
	var isbn, coverImage, source, dateAcquired *string
	err := r.db.QueryRow(
		`SELECT format, page_count, language, isbn, cover_image, source, date_acquired, second_hand
         FROM book_copy WHERE id = ?`,
		copyID,
	).Scan(&in.Format, &pageCount, &in.Language, &isbn, &coverImage, &source, &dateAcquired, &secondHand)
	if err != nil {
		return nil, err
	}
	if pageCount != nil {
		in.PageCount = *pageCount
	}
	in.ISBN = deref(isbn)
	in.CoverImage = deref(coverImage)
	in.Source = deref(source)
	in.DateAcquired = deref(dateAcquired)
	in.SecondHand = secondHand != nil && *secondHand == 1
	return &in, nil
}

func (r *Repo) UpdateCopy(copyID int, in *CopyInput) error {
	_, err := r.db.Exec(
		`UPDATE book_copy
         SET format = ?, page_count = ?, language = ?, isbn = ?, cover_image = ?, source = ?, date_acquired = ?, second_hand = ?
         WHERE id = ?`,
		in.Format,
		nullableInt(in.PageCount),
		in.Language,
		nullable(in.ISBN),
		nullable(in.CoverImage),
		nullable(in.Source),
		nullable(in.DateAcquired),
		boolInt(in.SecondHand),
		copyID,
	)
	return err
}

type Copy struct {
	ID     int
	Format string
}

func (r *Repo) ListCopies(bookID int) ([]Copy, error) {
	rows, err := r.db.Query(
		"SELECT id, format FROM book_copy WHERE book_id = ? ORDER BY id",
		bookID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var copies []Copy
	for rows.Next() {
		var c Copy
		if err := rows.Scan(&c.ID, &c.Format); err != nil {
			return nil, err
		}
		copies = append(copies, c)
	}
	return copies, nil
}
