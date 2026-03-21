package repo

import "strings"

type ExportRead struct {
	// Book
	Title            string
	Headline         string
	BookType         string
	DatePublished    string
	OriginalLanguage string
	URL              string

	// Genre
	Genre string

	// Author (first only)
	Author string

	// Series
	SeriesName     string
	SeriesPosition float64

	// Copy
	Format     string
	Language   string
	PageCount  int
	CoverImage string

	// Read
	Status       string
	Rating       int
	DateStarted  string
	DateFinished string

	// Tags
	Tags []string
}

func (r *Repo) ListExportReads() ([]ExportRead, error) {
	rows, err := r.db.Query(
		`SELECT rd.id, rd.book_id, rd.status, rd.rating, rd.date_started, rd.date_finished,
                b.title, b.headline, b.type, b.date_published, b.original_language, b.url,
                b.series_id, b.series_position,
                g.name,
                bc.format, bc.language, bc.page_count, bc.cover_image
         FROM read rd
         JOIN book b ON b.id = rd.book_id
         JOIN genre g ON g.id = b.genre_id
         LEFT JOIN book_copy bc ON bc.id = rd.copy_id
         ORDER BY rd.date_finished DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type readRow struct {
		ExportRead
		readID int
		bookID int
	}

	var intermediate []readRow
	for rows.Next() {
		var rr readRow
		var headline, url, dateStarted, dateFinished *string
		var rating, pageCount *int
		var seriesID *int
		var seriesPosition *float64
		var format, copyLanguage, coverImage *string

		if err := rows.Scan(
			&rr.readID, &rr.bookID, &rr.ExportRead.Status, &rating, &dateStarted, &dateFinished,
			&rr.ExportRead.Title, &headline, &rr.ExportRead.BookType, &rr.ExportRead.DatePublished, &rr.ExportRead.OriginalLanguage, &url,
			&seriesID, &seriesPosition,
			&rr.ExportRead.Genre,
			&format, &copyLanguage, &pageCount, &coverImage,
		); err != nil {
			return nil, err
		}

		if headline != nil {
			rr.Headline = *headline
		}
		if url != nil {
			rr.URL = *url
		}
		if dateStarted != nil {
			rr.DateStarted = *dateStarted
		}
		if dateFinished != nil {
			rr.DateFinished = *dateFinished
		}
		if rating != nil {
			rr.Rating = *rating
		}
		if format != nil {
			rr.Format = *format
		}
		if copyLanguage != nil {
			rr.Language = *copyLanguage
		}
		if pageCount != nil {
			rr.PageCount = *pageCount
		}
		if coverImage != nil {
			rr.CoverImage = *coverImage
		}
		if seriesPosition != nil {
			rr.SeriesPosition = *seriesPosition
		}

		intermediate = append(intermediate, rr)
	}

	reads := make([]ExportRead, len(intermediate))
	for i, rr := range intermediate {
		reads[i] = rr.ExportRead

		// First author
		var author string
		err := r.db.QueryRow(
			`SELECT a.name FROM author a
             JOIN book_author ba ON ba.author_id = a.id
             WHERE ba.book_id = ? AND ba.role = 'author'
             LIMIT 1`,
			rr.bookID,
		).Scan(&author)
		if err == nil {
			reads[i].Author = author
		}

		// Series name
		if rr.SeriesPosition > 0 {
			var seriesName string
			err := r.db.QueryRow(
				`SELECT s.name FROM series s
                 JOIN book b ON b.series_id = s.id
                 WHERE b.id = ?`,
				rr.bookID,
			).Scan(&seriesName)
			if err == nil {
				reads[i].SeriesName = seriesName
			}
		}

		// Tags
		tagRows, err := r.db.Query(
			`SELECT t.name FROM tag t
             JOIN book_tag bt ON bt.tag_id = t.id
             WHERE bt.book_id = ?
             ORDER BY t.name`,
			rr.bookID,
		)
		if err == nil {
			for tagRows.Next() {
				var tag string
				if err := tagRows.Scan(&tag); err == nil {
					reads[i].Tags = append(reads[i].Tags, tag)
				}
			}
			tagRows.Close()
		}
	}

	return reads, nil
}

func (e *ExportRead) Slug() string {
	name := strings.ToLower(e.Title)
	name = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return ' '
	}, name)
	parts := strings.Fields(name)
	return strings.Join(parts, "-")
}

func (e *ExportRead) RatingDisplay() string {
	ratings := map[int]string{
		10: "5", 9: "4.5", 8: "4", 7: "3.5", 6: "3",
		5: "2.5", 4: "2", 3: "1.5", 2: "1", 1: "0.5",
	}
	if v, ok := ratings[e.Rating]; ok {
		return v
	}
	return ""
}
