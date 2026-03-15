package repo

type Author struct {
	ID       int
	Name     string
	SortName string
}

func (r *Repo) ListAuthors() ([]Author, error) {
	rows, err := r.db.Query("SELECT id, name, sort_name FROM author ORDER BY sort_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var authors []Author
	for rows.Next() {
		var a Author
		if err := rows.Scan(&a.ID, &a.Name, &a.SortName); err != nil {
			return nil, err
		}
		authors = append(authors, a)
	}
	return authors, nil
}
