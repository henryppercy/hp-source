package repo

import "fmt"

// CollectionBook is one book on a collection page (an author's, a tag's, a
// genre's or a series's), its reads folded into one representative status.
type CollectionBook struct {
	BookID        int
	Title         string
	CoverImage    string
	Status        string // reading | finished | abandoned | unread
	Rating        int
	DateStarted   string
	DateFinished  string
	CurrentPage   int
	PageCount     int
	DatePublished string
	DateAcquired  string
}

// ListBooksByAuthor returns the author's books that have a copy or a read,
// in publication order earliest first, each with its oldest copy merged in.
func (r *Repo) ListBooksByAuthor(authorID int) ([]CollectionBook, error) {
	rows, err := r.db.Query(
		`SELECT b.id, COALESCE(bc.title, b.title), bc.cover_image, rd.status, rd.rating,
		        rd.date_started, rd.date_finished,
		        (SELECT rl.page FROM read_log rl
		         WHERE rl.read_id = rd.id AND rl.page IS NOT NULL
		         ORDER BY rl.id DESC LIMIT 1),
		        bc.page_count, b.date_published, bc.date_acquired
		 FROM book b
		 JOIN book_author ba ON ba.book_id = b.id AND ba.role = 'author'
		 LEFT JOIN book_copy bc ON bc.id = (SELECT MIN(bc2.id) FROM book_copy bc2 WHERE bc2.book_id = b.id)
		 LEFT JOIN read rd ON rd.id = (
		     SELECT r2.id FROM read r2 WHERE r2.book_id = b.id
		     ORDER BY CASE r2.status WHEN 'reading' THEN 0 WHEN 'finished' THEN 1 ELSE 2 END,
		              r2.date_finished DESC, r2.id DESC
		     LIMIT 1)
		 WHERE ba.author_id = ?
		   AND (EXISTS (SELECT 1 FROM book_copy bc3 WHERE bc3.book_id = b.id)
		        OR EXISTS (SELECT 1 FROM read r3 WHERE r3.book_id = b.id))
		 ORDER BY b.date_published, b.title, b.id`,
		authorID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list books by author: %w", err)
	}
	defer rows.Close()

	var books []CollectionBook
	for rows.Next() {
		var b CollectionBook
		var coverImage, status, dateStarted, dateFinished *string
		var rating, pageCount, currentPage *int
		var dateAcquired *string
		if err := rows.Scan(
			&b.BookID, &b.Title, &coverImage, &status, &rating,
			&dateStarted, &dateFinished, &currentPage, &pageCount,
			&b.DatePublished, &dateAcquired,
		); err != nil {
			return nil, fmt.Errorf("failed to scan collection book: %w", err)
		}
		b.CoverImage = deref(coverImage)
		if status != nil {
			b.Status = *status
		} else {
			b.Status = "unread"
		}
		if rating != nil {
			b.Rating = *rating
		}
		b.DateStarted = deref(dateStarted)
		b.DateFinished = deref(dateFinished)
		if currentPage != nil {
			b.CurrentPage = *currentPage
		}
		if pageCount != nil {
			b.PageCount = *pageCount
		}
		b.DateAcquired = deref(dateAcquired)
		books = append(books, b)
	}
	return books, rows.Err()
}
