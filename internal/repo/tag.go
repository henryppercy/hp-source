package repo

type Tag struct {
	ID   int
	Name string
}

func (r *Repo) ListTags() ([]Tag, error) {
	rows, err := r.db.Query("SELECT id, name FROM tag ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}
