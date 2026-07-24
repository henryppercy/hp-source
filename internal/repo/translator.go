package repo

type Translator struct {
	ID       int
	Name     string
	SortName string
}

type TranslatorInput struct {
	ID       int
	Name     string
	SortName string
}

func (r *Repo) ListTranslators() ([]Translator, error) {
	rows, err := r.db.Query("SELECT id, name, sort_name FROM translator ORDER BY sort_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translators []Translator
	for rows.Next() {
		var t Translator
		if err := rows.Scan(&t.ID, &t.Name, &t.SortName); err != nil {
			return nil, err
		}
		translators = append(translators, t)
	}
	return translators, nil
}

func (r *Repo) CreateTranslator(name, sortName string) (int, error) {
	return createTranslator(r.db, name, sortName)
}

func createTranslator(tx TX, name, sortName string) (int, error) {
	result, err := tx.Exec(
		"INSERT INTO translator (name, sort_name) VALUES (?, ?)",
		name, sortName,
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

func linkCopyTranslator(tx TX, copyID, translatorID int) error {
	_, err := tx.Exec(
		"INSERT INTO copy_translator (copy_id, translator_id) VALUES (?, ?)",
		copyID, translatorID,
	)
	return err
}

// loadCopyTranslators returns the translators linked to a copy, for prefilling
// the edit form.
func loadCopyTranslators(tx TX, copyID int) ([]TranslatorInput, error) {
	rows, err := tx.Query(
		`SELECT t.id, t.name, t.sort_name FROM translator t
         JOIN copy_translator ct ON ct.translator_id = t.id
         WHERE ct.copy_id = ? ORDER BY t.sort_name`,
		copyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var translators []TranslatorInput
	for rows.Next() {
		var t TranslatorInput
		if err := rows.Scan(&t.ID, &t.Name, &t.SortName); err != nil {
			return nil, err
		}
		translators = append(translators, t)
	}
	return translators, nil
}

// linkTranslators resolves each input to a translator id (creating new ones) and
// links them to the copy.
func linkTranslators(tx TX, copyID int, translators []TranslatorInput) error {
	for _, t := range translators {
		id := t.ID
		if id == 0 {
			var err error
			id, err = createTranslator(tx, t.Name, t.SortName)
			if err != nil {
				return err
			}
		}
		if err := linkCopyTranslator(tx, copyID, id); err != nil {
			return err
		}
	}
	return nil
}
