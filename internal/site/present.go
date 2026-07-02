package site

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/site/templates"
	"github.com/henryppercy/hp-source/internal/text"
)

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

func toListItem(p repo.Post) templates.PostListItem {
	return templates.PostListItem{
		Title:       p.Title,
		Slug:        p.Slug,
		URL:         postURL(p),
		PublishedAt: parseDate(p.PublishedAt),
		Headline:    p.Headline,
		Topics:      topicLinks(p.Topics),
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
func topicLinks(topics []repo.Topic) []templates.TopicLink {
	links := make([]templates.TopicLink, len(topics))
	for i, t := range topics {
		links[i] = templates.TopicLink{Name: t.Name, URL: "/topics/" + text.Slug(t.Name)}
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
func articleItems(posts []repo.Post, keep func(repo.Post) bool) []templates.PostListItem {
	var items []templates.PostListItem
	for _, p := range posts {
		if p.Kind == "article" && keep(p) {
			items = append(items, toListItem(p))
		}
	}
	return items
}

// mainArticles are the articles for /posts and the home stream: every article
// except the Spanish learning log (posts whose only topic is spanish, which live
// on /spanish). Posts that touch spanish among other topics still appear.
func mainArticles(posts []repo.Post) []templates.PostListItem {
	return articleItems(posts, func(p repo.Post) bool {
		return !onlySpanish(p)
	})
}

// onlySpanish reports whether spanish is the post's sole topic.
func onlySpanish(p repo.Post) bool {
	return len(p.Topics) == 1 && p.Topics[0].Name == "spanish"
}

// articlesWithTopic maps the articles carrying the named topic to list items.
func articlesWithTopic(posts []repo.Post, name string) []templates.PostListItem {
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

func booksReadInYear(reads []repo.ReadEntry, year int) int {
	count := 0
	for _, e := range reads {
		if e.Status == "finished" && parseDate(e.DateFinished).Year() == year {
			count++
		}
	}
	return count
}

// readingHub assembles the /reading hub: the desk, this year's almanac and log,
// the year nav, and a shelf summary.
func readingHub(reads []repo.ReadEntry, shelf []repo.ShelfEntry, year int) templates.ReadingView {
	log := finishedLog(reads)
	return templates.ReadingView{
		Almanac:   almanac(reads, year),
		Reading:   currentReads(reads),
		Year:      yearLog(log, year),
		SetAside:  abandonedInYear(reads, year),
		TotalRead: countStatus(reads, "finished"),
		Nav:       yearLinks(logYears(log), year, year),
		Insights:  yearInsights(reads, year),
		Shelf:     shelfSummary(shelfBooks(shelf)),
	}
}

// yearPage pairs a year with its rendered view so the builder can route it.
type yearPage struct {
	Year int
	View templates.YearView
}

// readingYearPages builds a recap page for each past year with finished books;
// the current year lives on the hub, not its own page.
func readingYearPages(reads []repo.ReadEntry, current int) []yearPage {
	log := finishedLog(reads)
	years := logYears(log)
	var pages []yearPage
	for _, y := range log {
		if y.Year == current {
			continue
		}
		pages = append(pages, yearPage{
			Year: y.Year,
			View: templates.YearView{
				Almanac:  almanac(reads, y.Year),
				Year:     y,
				SetAside: abandonedInYear(reads, y.Year),
				Insights: yearInsights(reads, y.Year),
				Nav:      yearLinks(years, current, y.Year),
			},
		})
	}
	return pages
}

func readingShelf(shelf []repo.ShelfEntry) templates.ShelfView {
	return templates.ShelfView{Books: shelfBooks(shelf)}
}

// yearLog returns the log for a single year, an empty year when none is found.
func yearLog(log []templates.FinishedYear, year int) templates.FinishedYear {
	for _, y := range log {
		if y.Year == year {
			return y
		}
	}
	return templates.FinishedYear{Year: year}
}

// logYears lists the years present in the log, newest first.
func logYears(log []templates.FinishedYear) []int {
	years := make([]int, 0, len(log))
	for _, y := range log {
		years = append(years, y.Year)
	}
	return years
}

// yearLinks builds the year nav from the years that have a page plus the current
// year. current routes to /reading; active marks the page being viewed.
func yearLinks(years []int, current, active int) []templates.YearLink {
	seen := map[int]bool{current: true}
	all := []int{current}
	for _, y := range years {
		if !seen[y] {
			seen[y] = true
			all = append(all, y)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(all)))
	links := make([]templates.YearLink, 0, len(all))
	for _, y := range all {
		url := fmt.Sprintf("/reading/%d", y)
		if y == current {
			url = "/reading"
		}
		links = append(links, templates.YearLink{
			Label:  fmt.Sprintf("%d", y),
			URL:    url,
			Active: y == active,
		})
	}
	return links
}

// abandonedInYear lists books set aside in the given year.
func abandonedInYear(reads []repo.ReadEntry, year int) []templates.FinishedBook {
	var out []templates.FinishedBook
	for _, e := range reads {
		if e.Status == "abandoned" && parseDate(e.DateFinished).Year() == year {
			out = append(out, finishedBook(e))
		}
	}
	return out
}

func shelfSummary(books []templates.ShelfBook) templates.ShelfSummary {
	s := templates.ShelfSummary{Total: len(books)}
	if len(books) > 0 {
		s.LongestWait = books[0].Waiting
		n := len(books)
		if n > 4 {
			n = 4
		}
		s.Oldest = books[:n]
	}
	return s
}

func countStatus(reads []repo.ReadEntry, status string) int {
	n := 0
	for _, e := range reads {
		if e.Status == status {
			n++
		}
	}
	return n
}

func currentReads(reads []repo.ReadEntry) []templates.CurrentRead {
	var out []templates.CurrentRead
	for _, e := range reads {
		if e.Status != "reading" {
			continue
		}
		started := parseDate(e.DateStarted)
		out = append(out, templates.CurrentRead{
			Title:     e.Title,
			Author:    e.Author,
			ImageURL:  coverURL(e.CoverImage),
			Format:    e.Format,
			StartedAt: started,
			DayCount:  daysSince(started),
			Percent:   readingPercent(e.CurrentPage, e.PageCount),
		})
	}
	return out
}

// readingPercent is the latest logged page against the copy's length, capped at
// 100 and zero until a page is logged.
func readingPercent(page, pages int) int {
	if page <= 0 || pages <= 0 {
		return 0
	}
	p := int(math.Round(float64(page) / float64(pages) * 100))
	if p > 100 {
		return 100
	}
	return p
}

// daysSince counts days from start to today, inclusive, zero when undated.
func daysSince(start time.Time) int {
	if start.IsZero() {
		return 0
	}
	d := int(time.Since(start).Hours()/24) + 1
	if d < 1 {
		return 1
	}
	return d
}

// finishedLog groups finished reads by the year they were finished, preserving
// the repo's newest-first order. Each book carries its lifetime index, counting
// down from the total so the newest book shows the running tally.
func finishedLog(reads []repo.ReadEntry) []templates.FinishedYear {
	n := 0
	for _, e := range reads {
		if e.Status == "finished" {
			n++
		}
	}
	var years []templates.FinishedYear
	at := map[int]int{}
	for _, e := range reads {
		if e.Status != "finished" {
			continue
		}
		b := finishedBook(e)
		b.Index = n
		n--
		y := b.FinishedAt.Year()
		i, ok := at[y]
		if !ok {
			i = len(years)
			at[y] = i
			years = append(years, templates.FinishedYear{Year: y})
		}
		years[i].Books = append(years[i].Books, b)
	}
	return years
}

func finishedBook(e repo.ReadEntry) templates.FinishedBook {
	return templates.FinishedBook{
		Title:      e.Title,
		Author:     e.Author,
		Genre:      e.Genre,
		Rating:     e.Rating,
		RatingText: repo.RatingDisplay(e.Rating),
		StartedAt:  parseDate(e.DateStarted),
		FinishedAt: parseDate(e.DateFinished),
		DaysToRead: daysBetween(parseDate(e.DateStarted), parseDate(e.DateFinished)),
		Pages:      e.PageCount,
		Format:     e.Format,
		Standout:   e.Rating >= 9,
	}
}

// shelfBooks orders the antilibrary by how long each book has waited, longest
// first; undated acquisitions sink to the end.
func shelfBooks(shelf []repo.ShelfEntry) []templates.ShelfBook {
	var out []templates.ShelfBook
	for _, e := range shelf {
		acquired := parseDate(e.DateAcquired)
		out = append(out, templates.ShelfBook{
			Title:      e.Title,
			Author:     e.Author,
			ImageURL:   coverURL(e.CoverImage),
			Genre:      e.Genre,
			Pages:      e.PageCount,
			Format:     e.Format,
			AcquiredAt: acquired,
			Waiting:    humanizeSince(acquired),
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].AcquiredAt, out[j].AcquiredAt
		if a.IsZero() != b.IsZero() {
			return !a.IsZero()
		}
		return a.Before(b)
	})
	return out
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

// almanac tallies the year's finished reads: counts, pages, mean rating (in
// stars), mean pace, and the fiction and second-hand splits. Abandoned is an
// all-time count.
func almanac(reads []repo.ReadEntry, year int) templates.AlmanacView {
	a := templates.AlmanacView{Year: year, AvgRating: "—", AvgPace: "—"}
	ratingSum, ratingN, paceSum, paceN := 0, 0, 0, 0
	fiction, secondHand, sourced := 0, 0, 0
	for _, e := range reads {
		if e.Status == "abandoned" {
			a.Abandoned++
		}
		finished := parseDate(e.DateFinished)
		if e.Status != "finished" || finished.Year() != year {
			continue
		}
		a.Books++
		a.Pages += e.PageCount
		m := int(finished.Month()) - 1
		a.Months[m]++
		if a.Months[m] > a.PeakMonth {
			a.PeakMonth = a.Months[m]
		}
		if e.BookType == "fiction" {
			fiction++
		}
		if e.Source != "" {
			sourced++
			if isSecondHand(e.Source) {
				secondHand++
			}
		}
		if e.Rating > 0 {
			ratingSum += e.Rating
			ratingN++
		}
		if d := daysBetween(parseDate(e.DateStarted), finished); d > 0 {
			paceSum += d
			paceN++
		}
	}
	if ratingN > 0 {
		a.AvgRating = fmt.Sprintf("%.1f", float64(ratingSum)/float64(ratingN)/2)
	}
	if paceN > 0 {
		a.AvgPace = fmt.Sprintf("%dd", int(math.Round(float64(paceSum)/float64(paceN))))
	}
	a.FictionPct = percent(fiction, a.Books)
	a.FictionNote = countNote(fiction, a.Books)
	a.SecondHandPct = percent(secondHand, sourced)
	a.SecondHandNote = countNote(secondHand, sourced)
	return a
}

// isSecondHand reports whether a copy came already lived-in rather than bought
// new. Gifted counts as second-hand here; move it if that's wrong.
func isSecondHand(source string) bool {
	switch source {
	case "second-hand", "borrowed":
		return true
	}
	return false
}

// percent formats n/total as a whole percent, "—" when there is no base.
func percent(n, total int) string {
	if total <= 0 {
		return "—"
	}
	return fmt.Sprintf("%d%%", int(math.Round(float64(n)/float64(total)*100)))
}

// countNote is the "n of total" caption under a split figure.
func countNote(n, total int) string {
	if total <= 0 {
		return ""
	}
	return fmt.Sprintf("%d of %d", n, total)
}

// minStandoutBooks is the fewest finished books a year needs before standouts
// and superlatives say anything; below it the comparisons are just noise.
const minStandoutBooks = 4

// topTallyN caps the top-genres and top-authors lists.
const topTallyN = 3

// yearInsights derives the rail highlights for a year: standout books and the
// top genres and authors, each empty when the year is too thin to be worth it.
func yearInsights(reads []repo.ReadEntry, year int) templates.Insights {
	var books []repo.ReadEntry
	for _, e := range reads {
		if e.Status == "finished" && parseDate(e.DateFinished).Year() == year {
			books = append(books, e)
		}
	}
	return templates.Insights{
		Standouts:  standouts(books),
		TopGenres:  topGenres(books),
		TopAuthors: topAuthors(books),
	}
}

// standouts picks the year's fastest and slowest reads by pace (pages a day, so
// book length does not decide it) and the longest and shortest books, only once
// there are enough to compare and skipping any missing the pages or dates.
func standouts(books []repo.ReadEntry) []templates.Superlative {
	if len(books) < minStandoutBooks {
		return nil
	}
	var fastest, slowest, longest, shortest *repo.ReadEntry
	var fastPace, slowPace, longPages, shortPages int
	for i := range books {
		e := &books[i]
		days := daysBetween(parseDate(e.DateStarted), parseDate(e.DateFinished))
		if e.PageCount > 0 && days > 0 {
			pace := int(math.Round(float64(e.PageCount) / float64(days)))
			if pace < 1 {
				pace = 1
			}
			if fastest == nil || pace > fastPace {
				fastest, fastPace = e, pace
			}
			if slowest == nil || pace < slowPace {
				slowest, slowPace = e, pace
			}
		}
		if e.PageCount > 0 {
			if longest == nil || e.PageCount > longPages {
				longest, longPages = e, e.PageCount
			}
			if shortest == nil || e.PageCount < shortPages {
				shortest, shortPages = e, e.PageCount
			}
		}
	}
	var out []templates.Superlative
	if fastest != nil {
		out = append(out, superlative("fastest pace", fastest, fmt.Sprintf("%d pages/day", fastPace)))
	}
	if slowest != nil && slowest != fastest {
		out = append(out, superlative("slowest pace", slowest, fmt.Sprintf("%d pages/day", slowPace)))
	}
	if longest != nil {
		out = append(out, superlative("longest read", longest, fmt.Sprintf("%d pages", longPages)))
	}
	if shortest != nil && shortest != longest {
		out = append(out, superlative("shortest read", shortest, fmt.Sprintf("%d pages", shortPages)))
	}
	return out
}

func superlative(label string, e *repo.ReadEntry, value string) templates.Superlative {
	return templates.Superlative{
		Label:    label,
		Title:    e.Title,
		Author:   e.Author,
		ImageURL: coverURL(e.CoverImage),
		Value:    value,
	}
}

// topGenres ranks the year's genres by count, most read first, keeping the top
// few. It returns nothing when a single genre covers everything.
func topGenres(books []repo.ReadEntry) []templates.Tally {
	genres := tally(books, func(e repo.ReadEntry) string { return e.Genre })
	if len(genres) < 2 {
		return nil
	}
	return topTally(genres)
}

// topAuthors ranks the authors read more than once this year, most read first.
// A single read is the norm, so those are left out as uninteresting.
func topAuthors(books []repo.ReadEntry) []templates.Tally {
	authors := tally(books, func(e repo.ReadEntry) string { return e.Author })
	repeats := authors[:0]
	for _, a := range authors {
		if a.Count > 1 {
			repeats = append(repeats, a)
		}
	}
	return topTally(repeats)
}

// tally counts books by the key, preserving first-seen order and skipping blanks.
func tally(books []repo.ReadEntry, key func(repo.ReadEntry) string) []templates.Tally {
	counts := map[string]int{}
	var order []string
	for _, e := range books {
		k := key(e)
		if k == "" {
			continue
		}
		if counts[k] == 0 {
			order = append(order, k)
		}
		counts[k]++
	}
	out := make([]templates.Tally, 0, len(order))
	for _, k := range order {
		out = append(out, templates.Tally{Name: k, Count: counts[k]})
	}
	return out
}

// topTally sorts a tally by count, most first, and keeps the top few.
func topTally(t []templates.Tally) []templates.Tally {
	sort.SliceStable(t, func(i, j int) bool { return t[i].Count > t[j].Count })
	if len(t) > topTallyN {
		t = t[:topTallyN]
	}
	return t
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
