package templates

import (
	"fmt"
	"html/template"
	"strconv"
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

// readLabel is the post's reading-time stamp, e.g. "6 min read".
func readLabel(mins int) string {
	return strconv.Itoa(mins) + " min read"
}

// wordLabel is the post's length stamp with thousands commas, e.g. "1,240 words".
func wordLabel(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s + " words"
}

// BuildInfo is the build stamp shown in the header and footer, recording when
// the site was built, on what, and the place it was filed from. The builder
// sets LastBuild.
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

// editedLater reports whether updated falls on a later calendar day than
// published. That is the signal that a post was genuinely revised after going
// public rather than just created and published, which bumps updated_at the
// same day.
func editedLater(published, updated time.Time) bool {
	if published.IsZero() || updated.IsZero() {
		return false
	}
	py, pm, pd := published.Date()
	uy, um, ud := updated.Date()
	publishedDay := time.Date(py, pm, pd, 0, 0, 0, 0, time.UTC)
	updatedDay := time.Date(uy, um, ud, 0, 0, 0, 0, time.UTC)
	return updatedDay.After(publishedDay)
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
	ReadMinutes int
	Words       int
	TOC         []TOCEntry
	Topics      []TopicLink
	Location    Place
}

// Place is a location stamp for the location component. Coords is preformatted
// (e.g. "53.22°N 1.28°W"). Code and Coords are optional.
type Place struct {
	Name   string
	Code   string
	Coords string
}

// HomeLocation is the site's filed-from place, shown in the header, footer
// and home nameplate. The builder sets it from the database. The default keeps
// a standalone render sensible.
var HomeLocation = Place{Name: "Sheffield", Code: "GB", Coords: "53.22°N 1.28°W"}

// TOCEntry is a table-of-contents node. Children holds nested sub-headings.
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
	ReadMinutes int
	Topics      []TopicLink
	Location    Place
}

// HomeView is the homepage, from the latest-activity cells and the margin's
// recently-read shelf and Spanish streak ribbon to the merged stream and the
// section index. The nameplate and bio prose are static and live in the
// template.
type HomeView struct {
	Copy     HomeCopy
	Latest   []LatestCell
	Reads    []RecentRead
	Levels   DSLevelsView
	Subjects []TopicCount
	Stream   []FeedEntry
	Index    []IndexRow
}

// RecentRead is one cover in the home rail's recently-read shelf. Meta is
// the hover detail, the title, author and pace. URL links into the reading
// section.
type RecentRead struct {
	Title    string
	ImageURL string
	URL      string
	Meta     string
}

// DSLevelsView is the rail's Dreaming Spanish roadmap, the running total
// over a list of hour milestones, each dated when reached or pending.
type DSLevelsView struct {
	Head   string
	Figure string
	Levels []DSLevel
}

// DSLevel is one milestone in the roadmap, its hour threshold, and either
// the date it was reached (with a link to the related post, where one exists)
// or the hours still to go. Next marks the immediate target, kept in solid
// weight.
type DSLevel struct {
	Label   string // "600h"
	Reached bool
	Date    string // "14 Mar 2024" when reached
	URL     string // related post, reached milestones only
	ToGo    string // "600h to go" for every unreached milestone
	Next    bool   // the immediate next target: kept solid, the rest ghosted
}

// TopicCount is one subject in the margin's subjects card, a topic, its
// feed, and how many posts carry it.
type TopicCount struct {
	Name  string
	URL   string
	Count int
}

// LatestCell is one cell of the homepage's latest-activity strip, a
// freshest-fact pull from a single feed. Italic sets the lead in serif italic,
// for a book title. Sub is an optional muted line under the lead. URL, when
// set, makes the cell a link into its section.
type LatestCell struct {
	Kicker string
	Lead   string
	Sub    string
	Italic bool
	Meta   string
	URL    string
}

// FeedEntry is one item in the home stream, a union over the feeds. Kind
// selects how it renders, and only the fields that kind needs are set. Note
// carries the author (book) or headline (post). Rating is the book's 0-10
// score. ImageURL is its cover and Meta its genre, pages and days line.
// BodyHTML is the rendered note. Topics tag a post or note.
type FeedEntry struct {
	Kind      string // "post" | "note" | "book" | "milestone"
	Kicker    string
	Date      time.Time
	Title     string
	Note      string
	AuthorURL string // the author's page, when the note names one
	URL       string
	Rating    int
	ImageURL  string
	Meta      string
	BodyHTML  template.HTML
	Topics    []TopicLink
}

// IndexRow is one line of the section index, a numbered directory entry
// with a live count. URL is empty for a section that is not built yet.
type IndexRow struct {
	Num   string
	Label string
	URL   string
	Note  string
}

// PostListView is the top-level data for a post listing page (e.g. /posts).
type PostListView struct {
	Heading    string
	Subheading string
	Empty      string
	Posts      []PostListItem
}

// TopicFeedView is a topic page such as /topics/{topic} and /spanish, a
// list of articles and a timeline of slices, each rendered only when non-empty.
type TopicFeedView struct {
	Heading    string
	Subheading string
	Empty      string
	Articles   []PostListItem
	Slices     []SliceItem
}

// SpanishView is the /spanish page, the input dashboard above the writing
// feed.
type SpanishView struct {
	Total      string // lifetime hours, e.g. "472"
	Subheading string // the page lead under the title
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
	Records    []Stat
	Averages   []Stat
	Note       string // hand-written "where I'm at" line
	Articles   []PostListItem
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
	Verdict      string // finish vs deadline, one of "est. 4mo early", "est. on time", "reached", "stalled"
	Pace         string // "1h 3m a day lately, 54m a day needed"
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

// YearSpan is a year marker above the heatmap. Weeks is how many columns
// it covers, used to size its label.
type YearSpan struct {
	Label string
	Weeks int
}

// CalWeek is a Monday-to-Sunday column in the heatmap.
type CalWeek struct {
	Days [7]CalDay
}

// CalDay is one cell. InRange is false for padding before the first day or
// after today. Class is the shade for its logged time, Title the hover label.
type CalDay struct {
	InRange bool
	Class   string
	Title   string
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
	Heading    string
	Subheading string
	Empty      string
	Slices     []SliceItem
}

// EverythingFeedView is the top-level data for the /everything timeline, the
// complete merged stream of every feed. Figures counts the entries by kind
// for the header, Entries is the stream newest first.
type EverythingFeedView struct {
	Heading    string
	Subheading string
	Empty      string
	Figures    []Stat
	Entries    []FeedEntry
}

// ReadingView is the /reading hub, what's open now, this year's almanac
// and log, links to earlier years, the rail insights, and a shelf summary.
type ReadingView struct {
	Subheading string
	DeskEmpty  string // shown when nothing is currently being read
	Almanac    AlmanacView
	Reading    []CurrentRead
	Year       FinishedYear
	SetAside   []FinishedBook
	Nav        []YearLink
	Insights   Insights
	Shelf      ShelfSummary
}

// YearView is a /reading/{year} recap, that year's almanac, log and
// insights, with the year nav to move between years.
type YearView struct {
	Subheading string
	Almanac    AlmanacView
	Year       FinishedYear
	SetAside   []FinishedBook
	Insights   Insights
	Nav        []YearLink
}

// Insights are a year's rail highlights, standout books and the most-read
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
	URL      string
	Title    string
	Author   string
	ImageURL string
	Value    string // "84 pages/day" | "912 pages"
}

// Tally is one name with a count, shared by the top-genres and top-authors
// cards. URL links the name to its page when one exists.
type Tally struct {
	Name  string
	URL   string
	Count int
}

// ShelfView is the full antilibrary at /reading/shelf.
type ShelfView struct {
	Subheading string
	Books      []ShelfBook
}

// CollectionView is the shared page for one author, tag, genre or series, a
// kind label over the name, the tally strip beneath it, and the book grid
// split into status sections, each hung from a marginal head.
type CollectionView struct {
	Kicker   string // "Author", "Tag", "Genre", "Series"
	Title    string
	Figures  []Stat
	Sections []CollectionSection
}

// CollectionSection is one status slice of a collection, reading, read, set
// aside or unread, its books in publication order earliest first.
type CollectionSection struct {
	Label string
	Books []CollectionBookView
}

// CollectionBookView is one book in a collection grid. Status decides the
// cell's boxed meta fields, the printed part in mono with one written-in
// serif italic value.
type CollectionBookView struct {
	URL        string
	Title      string
	Author     string // left empty when the collection itself is the author
	ImageURL   string
	Status     string // reading | finished | abandoned | unread
	Rating     int
	Percent    int
	FinishedAt time.Time
	DayCount   int
	AcquiredAt time.Time
	Published  string // the year, "" when unknown
	Pages      int    // the copy's length, 0 when unknown
}

// BookView is a book's own page, the work with its oldest copy merged in,
// above every read of it.
type BookView struct {
	Title         string
	Kind          string           // fiction or non-fiction
	OriginalTitle string           // the title in the original language, shown in parentheses
	Credits       []BookCreditView // the names on the work, linked when known
	Translators   string
	Series        string
	Tags          []string
	Facts         []Stat
	ImageURL      string
	Head          string // the reads count, "once" or "3 times"
	Reads         []BookReadView
	UnreadNote    string // replaces the reads for a book never started
}

// BookCreditView is one credited name on a book page, its role for anyone
// credited beyond authorship, and its page link when they have one.
type BookCreditView struct {
	Name string
	Role string
	URL  string
}

// BookReadView is one read on a book page. Edition names a copy title that
// differs from the page's. Percent, CurrentPage and DayCount are set only
// for a read in progress.
type BookReadView struct {
	Status      string
	Edition     string
	Rating      int
	StartedAt   time.Time
	FinishedAt  time.Time
	Days        int
	Pages       int
	Pace        string // pages a day for a finished read with both dates, else ""
	Format      string
	Language    string
	Source      string // bought, gifted, borrowed, etc
	SecondHand  bool
	Percent     int
	CurrentPage int
	DayCount    int
}

// ShelfSummary is the hub's glance at the shelf, a count, the longest
// wait, and the few books that have waited longest.
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

// monthNames are the density strip's hover labels, January to December.
var monthNames = [12]string{"January", "February", "March", "April", "May", "June",
	"July", "August", "September", "October", "November", "December"}

// hoursLabel renders seconds as hours and minutes for a hover label, e.g.
// "12h 30m", "45m", "0m".
func hoursLabel(sec int) string {
	h, m := sec/3600, (sec%3600+30)/60
	if m == 60 {
		h++
		m = 0
	}
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	default:
		return fmt.Sprintf("%dm", m)
	}
}

// commas groups a whole number with thousands separators, e.g. 2312 -> "2,312".
func commas(n int) string {
	s := strconv.Itoa(n)
	neg := ""
	if n < 0 {
		neg, s = "-", s[1:]
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return neg + s
}

// plural returns "book" or "books" to match the count.
func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}

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

// AlmanacView is the year's reading tally for the hub plate. AvgRating and
// AvgPace are display-ready, "—" when there is nothing to average. MonthsPages
// holds per-month pages read for the density strip, indexed January to December.
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
	MonthsPages    [12]int // pages read per month
	MonthsBooks    [12]int // books finished per month
	PeakMonth      int     // pages in the busiest month, for scaling the bars
}

// CurrentRead is a book open on the desk now. DayCount is days since
// started. Percent is progress through the book. CurrentPage over Pages is
// that progress in pages.
type CurrentRead struct {
	URL         string // the book's page
	Title       string
	Author      string
	AuthorURL   string // the author's page, "" without one
	ImageURL    string
	StartedAt   time.Time
	DayCount    int
	Percent     int
	CurrentPage int
	Pages       int
}

// FinishedBook is one entry in the reading log. Index is its lifetime number
// (oldest is 1). Rating is the raw 0-10 score for the pip meter, RatingText its
// display form. Standout marks a top rating for the margin mark. URL links to
// the book's page.
type FinishedBook struct {
	Index      int
	URL        string
	Title      string
	Author     string
	AuthorURL  string // the author's page, "" without one
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

// ShelfBook is an owned, unread book, the antilibrary still waiting.
// Waiting is
// how long it has sat unread, "" when the acquisition date is unknown. Series is
// preformatted with its position, e.g. "Wolf Hall #2".
type ShelfBook struct {
	URL        string // the book's page
	Title      string
	Author     string
	AuthorURL  string // the author's page, "" without one
	ImageURL   string
	Genre      string
	Series     string
	Source     string
	SecondHand bool
	Pages      int
	Format     string
	AcquiredAt time.Time
	Waiting    string
}

// Pip is one mark in a rating meter, empty, half or full.
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
