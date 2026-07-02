package site

import (
	"fmt"
	"strings"
	"time"

	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/site/templates"
)

// This file holds the presentation helpers shared across the section domains
// (home, posts, reading, spanish): dates, locations, cover paths and small
// formatting. Domain-specific presenters live in their own files.

// parseDate parses the date/datetime text the repo stores, returning the zero
// time when empty or unrecognised.
func parseDate(s string) time.Time {
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// locationStamps maps location ids to display stamps for the post and slice
// cards. The builder populates it before rendering; a missing id yields a blank
// stamp, so unlocated content simply shows no place.
var locationStamps map[int]templates.Place

// coverURL maps a stored cover filename to its served path, leaving an empty
// value empty so a missing cover renders no image.
func coverURL(file string) string {
	if file == "" {
		return ""
	}
	return imageBase + "/" + file
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// humanizeSince renders the span from t to now in the coarsest sensible unit,
// "" when t is missing or in the future.
func humanizeSince(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	days := int(time.Since(t).Hours() / 24)
	switch {
	case days < 0:
		return ""
	case days < 14:
		return fmt.Sprintf("%dd", days)
	case days < 60:
		return fmt.Sprintf("%dw", days/7)
	case days < 730:
		return fmt.Sprintf("%dmo", days/30)
	default:
		return fmt.Sprintf("%dyr", days/365)
	}
}

// daysBetween is calendar days from start to end counting both endpoints, so a
// book started and finished the same day took one day and consecutive days took
// two. Zero when either date is missing or end precedes start.
func daysBetween(start, end time.Time) int {
	if start.IsZero() || end.IsZero() {
		return 0
	}
	start = start.Truncate(24 * time.Hour)
	end = end.Truncate(24 * time.Hour)
	d := int(end.Sub(start).Hours()/24) + 1
	if d < 1 {
		return 0
	}
	return d
}

// HomeSlug is the location I live in, shown on the home nameplate and used as
// the default place a build is filed from.
const HomeSlug = "sheffield"

// placeOf turns a stored location into a display stamp, formatting its raw
// coordinates.
func placeOf(l repo.Location) templates.Place {
	return templates.Place{Name: l.Name, Code: l.Code, Coords: formatCoords(l.Lat, l.Lng)}
}

// formatCoords renders raw decimal degrees as "53.22°N 1.28°W", blank when
// either coordinate is missing.
func formatCoords(lat, lng *float64) string {
	if lat == nil || lng == nil {
		return ""
	}
	ns, la := "N", *lat
	if la < 0 {
		ns, la = "S", -la
	}
	ew, lo := "E", *lng
	if lo < 0 {
		ew, lo = "W", -lo
	}
	return fmt.Sprintf("%.2f°%s %.2f°%s", la, ns, lo, ew)
}
