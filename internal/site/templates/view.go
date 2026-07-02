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

// BuildInfo is the colophon shown in the header and footer: when the site was
// built, on what, and the place it was filed from. The builder sets LastBuild.
type BuildInfo struct {
	Date     time.Time
	Go       string
	On       string
	Location Place
}

// LastBuild drives the header date and the footer's "The Build" panel.
var LastBuild BuildInfo

// buildDate is the header masthead date, e.g. "Sat 14 Jun 2026".
func buildDate(t time.Time) string {
	return t.Format("Mon 2 Jan 2006")
}

// buildStamp is the footer's "last build" line, e.g. "14 June 2026; 09:14".
func buildStamp(t time.Time) string {
	return t.Format("2 January 2006; 15:04")
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
	Location    Place
}

// Place is a location stamp for the location component. Coords is preformatted
// (e.g. "53.22°N 1.28°W"); Code and Coords are optional.
type Place struct {
	Name   string
	Code   string
	Coords string
}

// HomeLocation is the site's filed-from place, shown in the header, footer and
// home nameplate. The builder sets it from the database; the default keeps a
// standalone render sensible.
var HomeLocation = Place{Name: "Sheffield", Code: "GB", Coords: "53.22°N 1.28°W"}

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
	Location    Place
}

// HomeView is the frontispiece: the dispatch strip's live cells, the colophon
// stat cells, the merged stream of everything, and the section index. The
// nameplate and bio prose are static and live in the template.
type HomeView struct {
	Dispatch []DispatchCell
	Stats    []Stat
	Subjects []TopicCount
	Stream   []FeedEntry
	Index    []IndexRow
}

// TopicCount is one subject in the margin's subjects card: a topic, its feed,
// and how many posts carry it.
type TopicCount struct {
	Name  string
	URL   string
	Count int
}

// DispatchCell is one cell of the dispatch strip: a freshest-fact pull from a
// single feed. Italic sets the lead in serif italic (for a book title). URL,
// when set, makes the cell a link into its section.
type DispatchCell struct {
	Kicker string
	Lead   string
	Italic bool
	Meta   string
	URL    string
}

// FeedEntry is one item in the home stream, a union over the feeds. Kind selects
// how it renders; only the fields that kind needs are set. Note carries the
// author (book) or headline (post); Rating is the book's 0-10 score; ImageURL
// its cover and Meta its "genre ; pages ; days" line; BodyHTML is the rendered
// note; Topics tag a post or note.
type FeedEntry struct {
	Kind     string // "post" | "note" | "book" | "milestone"
	Kicker   string
	Date     time.Time
	Title    string
	Note     string
	URL      string
	Rating   int
	ImageURL string
	Meta     string
	BodyHTML template.HTML
	Topics   []TopicLink
}

// IndexRow is one line of the section index: a numbered directory entry with a
// live count. URL is empty for a section that is not built yet.
type IndexRow struct {
	Num   string
	Label string
	URL   string
	Note  string
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

// SpanishView is the /spanish page: the input dashboard above the writing feed.
type SpanishView struct {
	Total      string // lifetime hours, e.g. "472"
	Intro      string // the serif interlude framing the daily log
	Band       BandView
	StartDate  time.Time
	DayCount   int
	Year       int
	Stats      []Stat  // the six frontispiece figures
	Months     [12]int // seconds logged per month, January to December
	PeakMonth  int     // seconds in the busiest month, for scaling the bars
	PeakLabel  string
	Goal       GoalView
	Calendar   CalendarView
	Milestones []MilestoneRung
	Records    []Stat
	Averages   []Stat
	Note       string // hand-written "where I'm at" line
	Articles   []PostListItem
	Slices     []SliceItem
}

// Stat is a label over or beside a figure, shared across the Spanish cards.
type Stat struct {
	Label string
	Value string
}

// BandView is progress through the current hours band toward the next mark, the
// hero's replacement for a named level. AtMax is set past the last mark.
type BandView struct {
	PrevLabel string // "300h"
	NextLabel string // "600h"
	Pct       int
	ToNext    string // "128 hours to 600"
	AtMax     bool
}

// GoalView is the burn-up toward the year's hour target. ActualPoints and
// PacePoints are SVG polyline coordinates in a 720x240 box.
type GoalView struct {
	Head         string // "800 hours by December 2026"
	Verdict      string // "14h ahead" | "reached"
	Delta        string // compact form for the stat cell: "+14h" | "-6h" | "on 800h"
	Pace         string // "1h 40m a day to finish on time"
	Reached      bool
	ActualPoints string
	PacePoints   string
	NowX         string
	NowY         string
}

// CalendarView is the contribution heatmap, one column per week from the first
// logged day to today. Years are the labels above it, each spanning its weeks.
type CalendarView struct {
	Weeks []CalWeek
	Years []YearSpan
}

// YearSpan is a year marker above the heatmap; Weeks is how many columns it
// covers, used to size its label.
type YearSpan struct {
	Label string
	Weeks int
}

// CalWeek is a Monday-to-Sunday column in the heatmap.
type CalWeek struct {
	Days [7]CalDay
}

// CalDay is one cell: InRange is false for padding before the first day or after
// today. Class is the shade for its logged time, Title the hover label.
type CalDay struct {
	InRange bool
	Class   string
	Title   string
}

// MilestoneRung is one rung of the roadmap ladder. URL is set only for a reached
// milestone that has a reflection post.
type MilestoneRung struct {
	Label   string
	Reached bool
	Date    time.Time
	URL     string
}

// SliceItem is one note in the /slices timeline, body rendered inline. Slug is
// the in-page anchor on the feed and the fragment a permalink links back to.
type SliceItem struct {
	URL         string
	Slug        string
	PublishedAt time.Time
	BodyHTML    template.HTML
	Topics      []TopicLink
	Location    Place
}

// SliceFeedView is the top-level data for the /slices timeline.
type SliceFeedView struct {
	Heading string
	Intro   string
	Slices  []SliceItem
}

// ReadingView is the /reading hub: what's open now, this year's almanac and log,
// links to earlier years, the rail insights, and a shelf summary.
type ReadingView struct {
	Almanac   AlmanacView
	Reading   []CurrentRead
	Year      FinishedYear
	SetAside  []FinishedBook
	TotalRead int
	Nav       []YearLink
	Insights  Insights
	Shelf     ShelfSummary
}

// YearView is a /reading/{year} recap: that year's almanac, log and insights,
// with the year nav to move between years.
type YearView struct {
	Almanac  AlmanacView
	Year     FinishedYear
	SetAside []FinishedBook
	Insights Insights
	Nav      []YearLink
}

// Insights are a year's rail highlights: standout books and the most-read
// genres and authors. Slices are empty when there is too little to be useful.
type Insights struct {
	Standouts  []Superlative
	TopGenres  []Tally
	TopAuthors []Tally
}

// hasInsights reports whether any rail card has content to show.
func hasInsights(i Insights) bool {
	return len(i.Standouts) > 0 || len(i.TopGenres) > 0 || len(i.TopAuthors) > 0
}

// Superlative is one standout book, e.g. the fastest read, with its cover.
type Superlative struct {
	Label    string // "fastest" | "slowest" | "longest" | "shortest"
	Title    string
	Author   string
	ImageURL string
	Value    string // "84 pages/day" | "912 pages"
}

// Tally is one name with a count, shared by the top-genres and top-authors cards.
type Tally struct {
	Name  string
	Count int
}

// ShelfView is the full antilibrary at /reading/shelf.
type ShelfView struct {
	Books []ShelfBook
}

// ShelfSummary is the hub's glance at the shelf: a count, the longest wait, and
// the few books that have waited longest.
type ShelfSummary struct {
	Total       int
	LongestWait string
	Oldest      []ShelfBook
}

// YearLink is one tab in the reading year nav. Active marks the current page.
type YearLink struct {
	Label  string
	URL    string
	Active bool
}

// monthInitials labels the almanac density strip, January to December.
var monthInitials = [12]string{"J", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"}

// barPct is a month's height in the density strip as a percent of the peak
// month, with a floor so a non-zero month still shows.
func barPct(count, peak int) int {
	if count <= 0 || peak <= 0 {
		return 0
	}
	if pct := count * 100 / peak; pct >= 12 {
		return pct
	}
	return 12
}

// AlmanacView is the year's reading tally for the frontispiece. AvgRating and
// AvgPace are display-ready, "—" when there is nothing to average. Months holds
// a per-month finished count for the density strip, indexed January to December.
type AlmanacView struct {
	Year           int
	Books          int
	Pages          int
	AvgRating      string
	AvgPace        string
	FictionPct     string // percent of the year that was fiction
	FictionNote    string // "17 of 19"
	SecondHandPct  string // percent that came second-hand
	SecondHandNote string
	Abandoned      int
	Months         [12]int
	PeakMonth      int
}

// CurrentRead is a book open on the desk now. DayCount is days since started;
// Percent is progress through the book.
type CurrentRead struct {
	Title     string
	Author    string
	ImageURL  string
	Format    string
	StartedAt time.Time
	DayCount  int
	Percent   int
}

// FinishedBook is one entry in the reading log. Index is its lifetime number
// (oldest is 1). Rating is the raw 0-10 score for the pip meter, RatingText its
// display form. Standout marks a top rating for the margin mark.
type FinishedBook struct {
	Index      int
	Title      string
	Author     string
	Genre      string
	Rating     int
	RatingText string
	StartedAt  time.Time
	FinishedAt time.Time
	DaysToRead int
	Pages      int
	Format     string
	Standout   bool
}

// FinishedYear groups finished books under the year they were finished.
type FinishedYear struct {
	Year  int
	Books []FinishedBook
}

// ShelfBook is an owned, unread book: the antilibrary still waiting. Waiting is
// how long it has sat unread, "" when the acquisition date is unknown.
type ShelfBook struct {
	Title      string
	Author     string
	ImageURL   string
	Genre      string
	Pages      int
	Format     string
	AcquiredAt time.Time
	Waiting    string
}

// Pip is one mark in a rating meter: empty, half or full.
type Pip int

const (
	PipEmpty Pip = iota
	PipHalf
	PipFull
)

// ratingPips maps a raw 0-10 rating to five pips, each a half-point of stars.
func ratingPips(rating int) []Pip {
	pips := make([]Pip, 5)
	for i := range pips {
		switch {
		case rating >= (i+1)*2:
			pips[i] = PipFull
		case rating == i*2+1:
			pips[i] = PipHalf
		default:
			pips[i] = PipEmpty
		}
	}
	return pips
}
