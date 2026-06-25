package site

import (
	"html/template"
	"time"
)

// PostView is a single article rendered at /posts/{slug}.
type PostView struct {
	Title       string
	Slug        string
	Type        string // "" | "slice" | "spanish"
	PublishedAt time.Time
	UpdatedAt   time.Time
	Headline    string
	BodyHTML    template.HTML
	TOC         []TOCEntry
}

// TOCEntry is a table-of-contents node; Children holds nested sub-headings.
type TOCEntry struct {
	Title    string
	Anchor   string
	Children []TOCEntry
}

// PostListItem is a row in the /posts, /spanish and /slices listings.
type PostListItem struct {
	Title       string
	Slug        string
	Type        string
	URL         string
	PublishedAt time.Time
	Headline    string
}

// BookView is a book on /reading and in the home recent-books pull.
type BookView struct {
	Title        string
	Author       string
	ImageURL     string
	Status       string // "reading" | "finished"
	Rating       string // display form, e.g. "4.5"
	DateFinished time.Time
}

// HomeView is the top-level data for the home page.
type HomeView struct {
	RecentBooks     []BookView
	RecentPosts     []PostListItem
	BooksReadInYear int
	Year            int
}

// PostListView is the top-level data for a post listing page. Reused by
// /posts, /spanish and /slices.
type PostListView struct {
	Heading string
	Posts   []PostListItem
}

// ReadingView is the top-level data for the /reading page.
type ReadingView struct {
	CurrentlyReading []BookView
	Finished         []BookView
}
