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
