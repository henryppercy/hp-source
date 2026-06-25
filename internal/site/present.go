package site

import (
	"time"

	"github.com/henryppercy/hp-source/internal/repo"
)

const recentLimit = 5

// parseDate parses the date/datetime text the repo stores, returning the zero
// time when empty or unrecognised.
func parseDate(s string) time.Time {
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02", time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func toListItem(p repo.Post) PostListItem {
	return PostListItem{
		Title:       p.Title,
		Slug:        p.Slug,
		Type:        p.Type,
		URL:         postURL(p),
		PublishedAt: parseDate(p.PublishedAt),
		Headline:    p.Headline,
	}
}

// postURL is a post's canonical page path. Slices live under /slices; posts and
// spanish entries both render under /posts.
func postURL(p repo.Post) string {
	if p.Type == "slice" {
		return "/slices/" + p.Slug
	}
	return "/posts/" + p.Slug
}

func listItemsByType(posts []repo.Post, typ string) []PostListItem {
	var items []PostListItem
	for _, p := range posts {
		if p.Type == typ {
			items = append(items, toListItem(p))
		}
	}
	return items
}

// recentPosts returns the n most recent real posts (type "") for the home page,
// excluding slices and spanish. Input is assumed newest-first; it slices, never
// sorts.
func recentPosts(posts []repo.Post, n int) []PostListItem {
	items := listItemsByType(posts, "")
	if len(items) > n {
		items = items[:n]
	}
	return items
}

func bookView(e repo.ReadEntry) BookView {
	return BookView{
		Title:        e.Title,
		Author:       e.Author,
		ImageURL:     e.CoverImage,
		Status:       e.Status,
		Rating:       repo.RatingDisplay(e.Rating),
		DateFinished: parseDate(e.DateFinished),
	}
}

func bookViews(reads []repo.ReadEntry) []BookView {
	var books []BookView
	for _, e := range reads {
		books = append(books, bookView(e))
	}
	return books
}

func booksByStatus(books []BookView, status string) []BookView {
	var out []BookView
	for _, b := range books {
		if b.Status == status {
			out = append(out, b)
		}
	}
	return out
}

// recentBooks returns up to n finished books, keeping the repo's
// newest-finished-first order; it filters then slices, never sorts.
func recentBooks(books []BookView, n int) []BookView {
	finished := booksByStatus(books, "finished")
	if len(finished) > n {
		finished = finished[:n]
	}
	return finished
}

func booksReadInYear(reads []repo.ReadEntry, year int) int {
	count := 0
	for _, e := range reads {
		if e.Status == "finished" && parseDate(e.DateFinished).Year() == year {
			count++
		}
	}
	return count
}
