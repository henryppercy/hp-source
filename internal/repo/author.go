package repo

type Author struct {
	ID       int
	Name     string
	SortName string
}

type AuthorInput struct {
	ID       int
	Name     string
	SortName string
	Role     string
}

func (r *Repo) ListAuthors() ([]Author, error) {
	return listAuthors(r.db)
}

func listAuthors(tx TX) ([]Author, error) {
	rows, err := tx.Query("SELECT id, name, sort_name FROM author ORDER BY sort_name")
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

func (r *Repo) CreateAuthor(name, sortName string) (int, error) {
	return createAuthor(r.db, name, sortName)
}

func createAuthor(tx TX, name, sortName string) (int, error) {
	result, err := tx.Exec(
		"INSERT INTO author (name, sort_name) VALUES (?, ?)",
		name, sortName,
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

func (r *Repo) LinkBookAuthor(bookID, authorID int, role string) error {
	return linkBookAuthor(r.db, bookID, authorID, role)
}

func linkBookAuthor(tx TX, bookID, authorID int, role string) error {
	_, err := tx.Exec(
		"INSERT INTO book_author (book_id, author_id, role) VALUES (?, ?, ?)",
		bookID, authorID, role,
	)
	return err
}
