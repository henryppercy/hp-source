package repo

import "fmt"

type BuildInput struct {
	LocationID int
	GoVersion  string
	BuiltOn    string
}

// AddBuild logs a build attempt as not-yet-successful and returns its id, to be
// flipped with MarkBuildSuccess once the build completes.
func (r *Repo) AddBuild(in *BuildInput) (int, error) {
	res, err := r.db.Exec(
		`INSERT INTO build (location_id, go_version, built_on) VALUES (?, ?, ?)`,
		nullableInt(in.LocationID), in.GoVersion, in.BuiltOn,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to add build: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get build id: %w", err)
	}
	return int(id), nil
}

func (r *Repo) MarkBuildSuccess(id int) error {
	if _, err := r.db.Exec(`UPDATE build SET success = 1 WHERE id = ?`, id); err != nil {
		return fmt.Errorf("failed to mark build success: %w", err)
	}
	return nil
}
