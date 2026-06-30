package templates

import (
	"html/template"
	"time"
)

// fmtDate formats a date for display, rendering the zero time as "" so missing
// dates show blank rather than a year-one placeholder.
func fmtDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2 Jan 2006")
}

// fmtDateTime formats a date with time of day, e.g. "14 Jun 2026, 14:32".
func fmtDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2 Jan 2006, 15:04")
}

// TopicLink is a topic shown on a page, linking to its feed.
type TopicLink struct {
	Name string
	URL  string
}

// PostView is a single article rendered at /posts/{slug}.
type PostView struct {
	Title       string
	Slug        string
	PublishedAt time.Time
	UpdatedAt   time.Time
	Headline    string
	BodyHTML    template.HTML
	TOC         []TOCEntry
	Topics      []TopicLink
}

// TOCEntry is a table-of-contents node; Children holds nested sub-headings.
type TOCEntry struct {
	Title    string
	Anchor   string
	Children []TOCEntry
}

// PostListItem is a row in the /posts, /spanish and topic listings.
type PostListItem struct {
	Title       string
	Slug        string
	URL         string
	PublishedAt time.Time
	Headline    string
	Topics      []TopicLink
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

// PostListView is the top-level data for a post listing page (e.g. /posts).
type PostListView struct {
	Heading string
	Posts   []PostListItem
}

// TopicFeedView is a topic page (/topics/{topic} and /spanish): a list of
// articles and a timeline of slices, each rendered only when non-empty.
type TopicFeedView struct {
	Heading  string
	Intro    string
	Articles []PostListItem
	Slices   []SliceItem
}

// SliceItem is one note in the /slices timeline, body rendered inline. Slug is
// the in-page anchor on the feed and the fragment a permalink links back to.
type SliceItem struct {
	URL         string
	Slug        string
	PublishedAt time.Time
	BodyHTML    template.HTML
	Topics      []TopicLink
}

// SliceFeedView is the top-level data for the /slices timeline.
type SliceFeedView struct {
	Heading string
	Intro   string
	Slices  []SliceItem
}

// ReadingView is the top-level data for the /reading page.
type ReadingView struct {
	CurrentlyReading []BookView
	Finished         []BookView
}
