package repo

import (
	"database/sql"
	"fmt"

	"github.com/henryppercy/hp-source/internal/text"
)

type Location struct {
	ID   int
	Slug string
	Name string
	Code string
	Lat  *float64
	Lng  *float64
}

type LocationInput struct {
	Name string
	Code string
	Slug string
	Lat  *float64
	Lng  *float64
}

func (r *Repo) ListLocations() ([]Location, error) {
	rows, err := r.db.Query(
		`SELECT id, slug, name, country_code, lat, lng FROM location ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}
	defer rows.Close()

	var locations []Location
	for rows.Next() {
		l, err := scanLocation(rows)
		if err != nil {
			return nil, err
		}
		locations = append(locations, l)
	}
	return locations, rows.Err()
}

func (r *Repo) GetLocationBySlug(slug string) (*Location, error) {
	rows, err := r.db.Query(
		`SELECT id, slug, name, country_code, lat, lng FROM location WHERE slug = ?`,
		slug,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("location %q not found", slug)
	}
	l, err := scanLocation(rows)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *Repo) CreateLocation(in *LocationInput) error {
	slug := text.Slug(in.Slug)
	if slug == "" {
		slug = text.Slug(in.Name)
	}
	_, err := r.db.Exec(
		`INSERT INTO location (slug, name, country_code, lat, lng) VALUES (?, ?, ?, ?, ?)`,
		slug, in.Name, nullable(in.Code), in.Lat, in.Lng,
	)
	if err != nil {
		return fmt.Errorf("failed to create location: %w", err)
	}
	return nil
}

func scanLocation(rows *sql.Rows) (Location, error) {
	var l Location
	var code *string
	if err := rows.Scan(&l.ID, &l.Slug, &l.Name, &code, &l.Lat, &l.Lng); err != nil {
		return Location{}, fmt.Errorf("failed to scan location: %w", err)
	}
	l.Code = deref(code)
	return l, nil
}
