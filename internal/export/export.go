package export

import (
	"fmt"
	"strings"

	"github.com/henryppercy/hp-source/internal/repo"
)

func MDX(e repo.ExportRead) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "---\n")
	fmt.Fprintf(&sb, "author: \"%s\"\n", e.Author)
	fmt.Fprintf(&sb, "title: \"%s\"\n", e.Title)
	fmt.Fprintf(&sb, "headline: \"%s\"\n", e.Headline)
	fmt.Fprintf(&sb, "series: \"%s\"\n", e.SeriesName)

	if e.SeriesPosition > 0 {
		fmt.Fprintf(&sb, "series_pos: %g\n", e.SeriesPosition)
	} else {
		fmt.Fprintf(&sb, "series_pos:\n")
	}

	fmt.Fprintf(&sb, "image_url: \"/%s\"\n", e.CoverImage)
	fmt.Fprintf(&sb, "date_published: %s\n", e.DatePublished)
	fmt.Fprintf(&sb, "date_started: %s\n", e.DateStarted)
	fmt.Fprintf(&sb, "date_finished: %s\n", e.DateFinished)

	if e.Rating > 0 {
		fmt.Fprintf(&sb, "rating: %s\n", repo.RatingDisplay(e.Rating))
	} else {
		fmt.Fprintf(&sb, "rating:\n")
	}

	fmt.Fprintf(&sb, "type: \"%s\"\n", e.BookType)
	fmt.Fprintf(&sb, "genre: \"%s\"\n", e.Genre)

	if len(e.Tags) > 0 {
		fmt.Fprintf(&sb, "tags:\n")
		for _, tag := range e.Tags {
			fmt.Fprintf(&sb, "  - \"%s\"\n", tag)
		}
	} else {
		fmt.Fprintf(&sb, "tags:\n")
	}

	fmt.Fprintf(&sb, "format: \"%s\"\n", e.Format)
	fmt.Fprintf(&sb, "language: \"%s\"\n", e.Language)
	fmt.Fprintf(&sb, "original_language: \"%s\"\n", e.OriginalLanguage)

	if e.PageCount > 0 {
		fmt.Fprintf(&sb, "page_count: %d\n", e.PageCount)
	} else {
		fmt.Fprintf(&sb, "page_count:\n")
	}

	fmt.Fprintf(&sb, "---\n")

	return sb.String()
}
