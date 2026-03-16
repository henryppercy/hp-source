package repo

import "fmt"

type BookInput struct {
	Title            string
	Headline         string
	BookType         string
	GenreID          int
	DatePublished    string
	OriginalLanguage string
	URL              string
	ShelfStatus      string

	Authors []AuthorInput
	Series  *SeriesInput
	TagIDs  []int

	Format       string
	PageCount    int
	Language     string
	ISBN         string
	CoverImage   string
	Source       string
	DateAcquired string
}

func (r *Repo) AddBook(input *BookInput) error {
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
		`INSERT INTO book (title, headline, date_published, original_language, type, genre_id, series_id, series_position, shelf_status, url)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Title,
		nullable(input.Headline),
		input.DatePublished,
		input.OriginalLanguage,
		input.BookType,
		input.GenreID,
		seriesID,
		seriesPosition,
		nullable(input.ShelfStatus),
		nullable(input.URL),
	)
	if err != nil {
		return fmt.Errorf("failed to create book: %w", err)
	}
	bookID, _ := result.LastInsertId()

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

	if input.ShelfStatus == "shelf" {
		_, err := createCopy(
			tx, int(bookID), input.Format,
			nullableInt(input.PageCount),
			input.Language, input.ISBN, input.CoverImage,
			input.Source, input.DateAcquired,
		)
		if err != nil {
			return fmt.Errorf("failed to create copy: %w", err)
		}
	}

	return tx.Commit()
}
