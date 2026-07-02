package forms

import (
	"fmt"
	"strconv"
	"strings"

	huh "charm.land/huh/v2"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/text"
)

// AddLocation collects a place: name, optional country code, optional
// coordinates, and a slug (defaulting to the slugified name).
func AddLocation(in *repo.LocationInput) error {
	var lat, lng string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Add Location").
				Next(true).
				NextLabel("Next"),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("required").
				Validate(huh.ValidateNotEmpty()).
				Value(&in.Name),
			huh.NewInput().
				Title("Country Code").
				Placeholder("optional, e.g. GB").
				Validate(validateCountryCode).
				Value(&in.Code),
			huh.NewInput().
				Title("Slug").
				PlaceholderFunc(func() string {
					return text.Slug(in.Name)
				}, &in.Name).
				Value(&in.Slug),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Latitude").
				Placeholder("optional, e.g. 53.22").
				Validate(validateOptionalFloat).
				Value(&lat),
			huh.NewInput().
				Title("Longitude").
				Placeholder("optional, e.g. -1.28").
				Validate(validateOptionalFloat).
				Value(&lng),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Confirm Location").
				DescriptionFunc(func() string {
					var sb strings.Builder
					fmt.Fprintf(&sb, "Name:       %s\n", in.Name)
					slug := text.Slug(in.Slug)
					if slug == "" {
						slug = text.Slug(in.Name)
					}
					fmt.Fprintf(&sb, "Slug:       %s\n", slug)
					if code := strings.ToUpper(strings.TrimSpace(in.Code)); code != "" {
						fmt.Fprintf(&sb, "Country:    %s\n", code)
					}
					if lat != "" && lng != "" {
						fmt.Fprintf(&sb, "Coords:     %s, %s\n", lat, lng)
					}
					return sb.String()
				}, &lng).
				Next(true).
				NextLabel("Save"),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	in.Code = strings.ToUpper(strings.TrimSpace(in.Code))
	in.Lat = parseFloatPtr(lat)
	in.Lng = parseFloatPtr(lng)
	return nil
}

func validateCountryCode(s string) error {
	if s == "" {
		return nil
	}
	if len(strings.TrimSpace(s)) != 2 {
		return fmt.Errorf("must be a 2-letter code, e.g. GB")
	}
	return nil
}

func validateOptionalFloat(s string) error {
	if s == "" {
		return nil
	}
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return fmt.Errorf("must be a number, e.g. 53.22")
	}
	return nil
}

func parseFloatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return &f
	}
	return nil
}
