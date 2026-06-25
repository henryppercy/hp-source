package site

import (
	"sort"
	"strings"
	"time"

	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/text"
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
		URL:         postURL(p),
		PublishedAt: parseDate(p.PublishedAt),
		Headline:    p.Headline,
	}
}

// postURL is a post's canonical page path. Slices live under /slices, articles
// under /posts.
func postURL(p repo.Post) string {
	if p.Kind == "slice" {
		return "/slices/" + p.Slug
	}
	return "/posts/" + p.Slug
}

// topicLinks maps a post's topics to display links to their feeds.
func topicLinks(topics []repo.Topic) []TopicLink {
	links := make([]TopicLink, len(topics))
	for i, t := range topics {
		links[i] = TopicLink{Name: t.Name, URL: "/topics/" + text.Slug(t.Name)}
	}
	return links
}

func hasTopic(p repo.Post, name string) bool {
	for _, t := range p.Topics {
		if t.Name == name {
			return true
		}
	}
	return false
}

// articleItems maps the articles matching keep to list items.
func articleItems(posts []repo.Post, keep func(repo.Post) bool) []PostListItem {
	var items []PostListItem
	for _, p := range posts {
		if p.Kind == "article" && keep(p) {
			items = append(items, toListItem(p))
		}
	}
	return items
}

// mainArticles are the articles for /posts and home recents: every article
// except those tagged spanish (which live on /spanish and /topics/spanish).
func mainArticles(posts []repo.Post) []PostListItem {
	return articleItems(posts, func(p repo.Post) bool {
		return !hasTopic(p, "spanish")
	})
}

// recentPosts returns the n most recent main articles for the home page. Input
// is assumed newest-first; it slices, never sorts.
func recentPosts(posts []repo.Post, n int) []PostListItem {
	items := mainArticles(posts)
	if len(items) > n {
		items = items[:n]
	}
	return items
}

// articlesWithTopic maps the articles carrying the named topic to list items.
func articlesWithTopic(posts []repo.Post, name string) []PostListItem {
	return articleItems(posts, func(p repo.Post) bool {
		return hasTopic(p, name)
	})
}

// slicesWithTopic returns the slice posts carrying the named topic.
func slicesWithTopic(posts []repo.Post, name string) []repo.Post {
	var out []repo.Post
	for _, p := range posts {
		if p.Kind == "slice" && hasTopic(p, name) {
			out = append(out, p)
		}
	}
	return out
}

// allSlices returns the slice posts, preserving order.
func allSlices(posts []repo.Post) []repo.Post {
	var out []repo.Post
	for _, p := range posts {
		if p.Kind == "slice" {
			out = append(out, p)
		}
	}
	return out
}

// usedTopics returns the distinct topic names present on the posts, sorted.
func usedTopics(posts []repo.Post) []string {
	seen := map[string]bool{}
	var names []string
	for _, p := range posts {
		for _, t := range p.Topics {
			if !seen[t.Name] {
				seen[t.Name] = true
				names = append(names, t.Name)
			}
		}
	}
	sort.Strings(names)
	return names
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// coverURL maps a stored cover filename to its served path, leaving an empty
// value empty so a missing cover renders no image.
func coverURL(file string) string {
	if file == "" {
		return ""
	}
	return imageBase + "/" + file
}

func bookView(e repo.ReadEntry) BookView {
	return BookView{
		Title:        e.Title,
		Author:       e.Author,
		ImageURL:     coverURL(e.CoverImage),
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
