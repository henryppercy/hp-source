package repo

func createCopy(tx TX, bookID int, format string, pageCount *int, language, isbn, coverImage, source, dateAcquired string) (int, error) {
	result, err := tx.Exec(
		`INSERT INTO book_copy (book_id, format, page_count, language, isbn, cover_image, source, date_acquired)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		bookID,
		format,
		pageCount,
		nullable(language),
		nullable(isbn),
		nullable(coverImage),
		nullable(source),
		nullable(dateAcquired),
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
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
