package repo

// Tag is a label for works.
type Tag struct {
	ID   int
	Name string
}

// ListTags returns every tag ordered by name.
func (r *Repo) ListTags() ([]Tag, error) {
	return listTags(r.db)
}

func listTags(tx TX) ([]Tag, error) {
	rows, err := tx.Query("SELECT id, name FROM tag ORDER BY name")
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

func linkBookTag(tx TX, bookID, tagID int) error {
	_, err := tx.Exec(
		"INSERT INTO book_tag (book_id, tag_id) VALUES (?, ?)",
		bookID, tagID,
	)
	return err
}
