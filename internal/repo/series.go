package repo

// Series is a named sequence of books.
type Series struct {
	ID   int
	Name string
}

// SeriesInput is a book's series membership, with Position its place in the
// sequence.
type SeriesInput struct {
	ID       int
	Name     string
	Position float64
}

// ListSeries returns every series ordered by name.
func (r *Repo) ListSeries() ([]Series, error) {
	return listSeries(r.db)
}

func listSeries(tx TX) ([]Series, error) {
	rows, err := tx.Query("SELECT id, name FROM series ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var series []Series
	for rows.Next() {
		var s Series
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			return nil, err
		}
		series = append(series, s)
	}
	return series, nil
}

// CreateSeries inserts a series and returns its new id.
func (r *Repo) CreateSeries(name string) (int, error) {
	return createSeries(r.db, name)
}

func createSeries(tx TX, name string) (int, error) {
	result, err := tx.Exec(
		"INSERT INTO series (name) VALUES (?)",
		name,
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}
