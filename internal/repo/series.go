package repo

type Series struct {
	ID   int
	Name string
}

func (r *Repo) ListSeries() ([]Series, error) {
	rows, err := r.db.Query("SELECT id, name FROM series ORDER BY name")
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
