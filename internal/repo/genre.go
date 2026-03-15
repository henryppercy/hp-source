package repo

type Genre struct {
	ID   int
	Name string
}

func (r *Repo) ListGenres() ([]Genre, error) {
	rows, err := r.db.Query("SELECT id, name FROM genre ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []Genre
	for rows.Next() {
		var g Genre
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	return genres, nil
}
