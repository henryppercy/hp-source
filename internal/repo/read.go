package repo

type ReadInput struct {
	BookID       int
	CopyID       int
	Status       string
	Rating       int
	DateStarted  string
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
