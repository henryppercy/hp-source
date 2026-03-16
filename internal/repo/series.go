package repo

type Series struct {
	ID   int
	Name string
}

type SeriesInput struct {
	ID       int
	Name     string
	Position float64
}

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
