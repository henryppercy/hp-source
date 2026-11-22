package repo

// Author is a writer credited on works.
type Author struct {
	ID       int
	Name     string
	SortName string
}

// AuthorInput is an author's editable fields. Role records how they
// contributed to a particular work.
type AuthorInput struct {
	ID       int
	Name     string
	SortName string
	Role     string
}

// ListAuthors returns every author ordered by sort name.
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

// CreateAuthor inserts an author and returns its new id.
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

// LinkBookAuthor credits an author on a work with the given role.
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
