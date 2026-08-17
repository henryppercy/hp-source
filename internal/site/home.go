package site

import (
	"fmt"
	"html"
	"html/template"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/site/templates"
	"github.com/henryppercy/hp-source/internal/text"
)

// homeStreamLimit caps the merged stream on the frontispiece.
const homeStreamLimit = 6

// homeShelfLimit caps the recently-read cover shelf in the rail
const homeShelfLimit = 12

// homeCurrently is the free-text "what I'm up to" line in the dispatch strip.
const homeCurrently = "Celebrating 500 hours of Spanish input."

// The frontispiece prose. The standfirst says who I am, the bio what the site
// is, so the two do not repeat each other.
const (
	homeKicker     = "Field notebook"
	homeHero       = "I like to keep a record."
	homeStandfirst = "Software developer based in Sheffield and semi-obsessive logger of things."
	homeBio        = "This is my digital field notebook; it's where I write down my thoughts, " +
		"log my reading, and track my Spanish learning. Everything enters via the command " +
		"line and the output is this website."
	homeStreamIntro = "This is everything: articles, notes, completed reads, spanish milestones; " +
		"being outputted in the reverse order that they came in."
)

// homeCopy is the frontispiece prose, gathered for the view.
var homeCopy = templates.HomeCopy{
	Kicker:      homeKicker,
	Hero:        homeHero,
	Standfirst:  homeStandfirst,
	Bio:         homeBio,
	StreamIntro: homeStreamIntro,
}

// homeView assembles the frontispiece: the dispatch strip, the colophon stats,
// the merged stream, and the section index.
func homeView(
	posts []repo.Post,
	reads []repo.ReadEntry,
	notes []templates.SliceItem,
	spanishLog []repo.SpanishLogEntry,
	now time.Time,
) templates.HomeView {
	articles := mainArticles(posts)

	days, _ := aggregateSpanish(spanishLog)
	total := 0
	for _, d := range days {
		total += d.sec
	}

	return templates.HomeView{
		Copy:     homeCopy,
		Dispatch: dispatchCells(reads, days, total, articles, notes, now),
		Reads:    recentReads(reads, homeShelfLimit),
		Levels:   spanishLevels(days, total, articles),
		Subjects: topicCounts(posts),
		Stream:   lifeStream(articles, notes, reads, spanishMilestones(days, total)),
		Index:    indexRows(articles, notes, reads, total),
	}
}

// topicCounts ranks the subjects across all posts by how many carry each, most
// first.
func topicCounts(posts []repo.Post) []templates.TopicCount {
	counts := map[string]int{}
	var order []string
	for _, p := range posts {
		for _, t := range p.Topics {
			if counts[t.Name] == 0 {
				order = append(order, t.Name)
			}
			counts[t.Name]++
		}
	}
	out := make([]templates.TopicCount, 0, len(order))
	for _, name := range order {
		out = append(out, templates.TopicCount{
			Name:  titleCase(name),
			URL:   "/topics/" + text.Slug(name),
			Count: counts[name],
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	return out
}

// dispatchCells build the dispatch strip: the freshest fact from each feed.
func dispatchCells(
	reads []repo.ReadEntry,
	days []spanishDay,
	total int,
	articles []templates.PostListItem,
	notes []templates.SliceItem,
	now time.Time,
) []templates.DispatchCell {
	var cells []templates.DispatchCell

	if cr := currentReads(reads); len(cr) > 0 {
		r := cr[0]
		meta := fmt.Sprintf("%d%%", r.Percent)
		if r.DayCount > 0 {
			meta += fmt.Sprintf("; day %d", r.DayCount)
		}
		kicker := "Open on the desk"
		var sub string
		if n := len(cr) - 1; n > 0 {
			kicker += fmt.Sprintf("; %d", len(cr))
			sub = fmt.Sprintf("+%d more open", n)
		}
		cells = append(cells, templates.DispatchCell{
			Kicker: kicker,
			Lead:   r.Title + ", " + r.Author,
			Sub:    sub,
			Italic: true,
			Meta:   meta,
			URL:    "/reading",
		})
	}

	if len(days) > 0 {
		cur, _ := streaks(days, dateOnly(now))
		dayCount := spanishDayCount(days, now)
		meta := fmt.Sprintf("%dh logged", total/3600)
		if cur > 0 {
			meta += fmt.Sprintf("; %dd streak", cur)
		}
		cells = append(cells, templates.DispatchCell{
			Kicker: "Spanish",
			Lead:   fmt.Sprintf("Day %d of comprehensible input", dayCount),
			Meta:   meta,
			URL:    "/spanish",
		})
	}

	if len(notes) > 0 {
		s := notes[0]
		cells = append(cells, templates.DispatchCell{
			Kicker: "Latest note",
			Lead:   teaser(s.BodyHTML, 90),
			Meta:   humanizeSince(s.PublishedAt),
			URL:    s.URL + "/",
		})
	}

	if len(articles) > 0 {
		p := articles[0]
		cells = append(cells, templates.DispatchCell{
			Kicker: "Latest post",
			Lead:   p.Title,
			Meta:   p.PublishedAt.Format("2 Jan 2006"),
			URL:    p.URL + "/",
		})
	}

	cells = append(cells, templates.DispatchCell{Kicker: "Currently", Lead: homeCurrently})
	return cells
}

// recentReads are the last n finished books for the rail's cover shelf, newest
// finish first.
func recentReads(reads []repo.ReadEntry, n int) []templates.RecentRead {
	var fin []repo.ReadEntry
	for _, e := range reads {
		if e.Status == "finished" {
			fin = append(fin, e)
		}
	}
	sort.SliceStable(fin, func(i, j int) bool {
		return parseDate(fin[i].DateFinished).After(parseDate(fin[j].DateFinished))
	})
	if len(fin) > n {
		fin = fin[:n]
	}
	out := make([]templates.RecentRead, 0, len(fin))
	for _, e := range fin {
		meta := e.Title
		if e.Author != "" {
			meta += ", " + e.Author
		}
		if e.Rating > 0 {
			meta += fmt.Sprintf("; %g/5", float64(e.Rating)/2)
		}
		if df := parseDate(e.DateFinished); !df.IsZero() {
			meta += "; finished " + df.Format("2 Jan 2006")
		}
		out = append(out, templates.RecentRead{
			Title:    e.Title,
			ImageURL: coverURL(e.CoverImage),
			URL:      "/reading",
			Meta:     meta,
		})
	}
	return out
}

// spanishLevels is the rail's Dreaming Spanish: each hour milestone with
// the date it was reached and a link to its post, or the hours still to go.
func spanishLevels(days []spanishDay, total int, articles []templates.PostListItem) templates.DSLevelsView {
	if len(days) == 0 {
		return templates.DSLevelsView{}
	}
	totalHours := total / 3600
	v := templates.DSLevelsView{
		Head:   "Spanish Milestones",
		Figure: fmt.Sprintf("%dh", totalHours),
	}
	nextMarked := false
	for _, t := range dsLevels {
		lvl := templates.DSLevel{Label: fmt.Sprintf("%dh", t)}
		if totalHours >= t {
			lvl.Reached = true
			lvl.Date = crossingDate(days, t).Format("02 Jan 2006")
			lvl.URL = milestonePost(articles, t)
		} else if !nextMarked {
			nextMarked = true
			lvl.Next = true
			lvl.ToGo = fmt.Sprintf("in %dh", t-totalHours)
		} else {
			lvl.ToGo = fmt.Sprintf("%dh", t-totalHours)
		}
		v.Levels = append(v.Levels, lvl)
	}
	return v
}

// milestonePost finds the article marking an hour milestone by the "-{hours}-hours"
// slug convention, returning its link or "" when none exists.
func milestonePost(articles []templates.PostListItem, hours int) string {
	suffix := fmt.Sprintf("-%d-hours", hours)
	for _, a := range articles {
		if strings.HasSuffix(a.Slug, suffix) {
			return a.URL + "/"
		}
	}
	return ""
}

// lifeStream folds the feeds into one reverse-chronological stream: articles,
// notes, finished books and crossed Spanish milestones, newest first.
func lifeStream(
	articles []templates.PostListItem,
	notes []templates.SliceItem,
	reads []repo.ReadEntry,
	milestones []templates.FeedEntry,
) []templates.FeedEntry {
	var out []templates.FeedEntry
	for _, p := range articles {
		out = append(out, templates.FeedEntry{
			Kind: "post", Kicker: "Filed", Date: p.PublishedAt,
			Title: p.Title, Note: p.Headline, URL: p.URL + "/", Topics: p.Topics,
		})
	}
	for _, s := range notes {
		out = append(out, templates.FeedEntry{
			Kind: "note", Kicker: "Note", Date: s.PublishedAt,
			URL: s.URL + "/", BodyHTML: s.BodyHTML, Topics: s.Topics,
		})
	}
	for _, e := range reads {
		if e.Status != "finished" {
			continue
		}
		out = append(out, templates.FeedEntry{
			Kind: "book", Kicker: "Finished", Date: parseDate(e.DateFinished),
			Title: e.Title, Note: e.Author, Rating: e.Rating,
			ImageURL: coverURL(e.CoverImage), Meta: bookMeta(e),
		})
	}
	out = append(out, milestones...)

	sort.SliceStable(out, func(i, j int) bool { return out[i].Date.After(out[j].Date) })
	if len(out) > homeStreamLimit {
		out = out[:homeStreamLimit]
	}
	return out
}

// bookMeta is a finished book's stream caption: genre, page count and how long
// it took, the parts that are known joined by semicolons.
func bookMeta(e repo.ReadEntry) string {
	var parts []string
	if e.Genre != "" {
		parts = append(parts, e.Genre)
	}
	if e.PageCount > 0 {
		parts = append(parts, fmt.Sprintf("%d pp", e.PageCount))
	}
	if d := daysBetween(parseDate(e.DateStarted), parseDate(e.DateFinished)); d > 0 {
		parts = append(parts, fmt.Sprintf("%d days", d))
	}
	return strings.Join(parts, "; ")
}

// spanishMilestones marks every 50 hours of Spanish input crossed, dated by the
// day the running total reached each step, for the home stream. Each carries the
// day of the journey it landed on and the pace since the previous mark.
func spanishMilestones(days []spanishDay, total int) []templates.FeedEntry {
	if len(days) == 0 {
		return nil
	}
	start := days[0].date
	prev := start
	var out []templates.FeedEntry
	for step := 50; step <= total/3600; step += 50 {
		date := crossingDate(days, step)
		segDays := daysBetween(prev, date)
		meta := fmt.Sprintf("50h in %d days", segDays)
		if segDays > 0 {
			meta += fmt.Sprintf("; %s avg per day", durShort(50*3600/segDays))
		}
		out = append(out, templates.FeedEntry{
			Kind: "milestone", Kicker: "Spanish", Date: date,
			Title: fmt.Sprintf("%s hours", commaNum(step)), URL: "/spanish",
			Meta: meta,
		})
		prev = date
	}
	return out
}

// indexRows build the section directory with a live count against each line.
func indexRows(
	articles []templates.PostListItem,
	notes []templates.SliceItem,
	reads []repo.ReadEntry,
	total int,
) []templates.IndexRow {
	return []templates.IndexRow{
		{Num: "02", Label: "Posts", URL: "/posts", Note: fmt.Sprintf("%d filed", len(articles))},
		{Num: "03", Label: "Slices", URL: "/slices", Note: fmt.Sprintf("%d notes", len(notes))},
		{Num: "04", Label: "Reading", URL: "/reading", Note: readingNote(reads)},
		{Num: "05", Label: "Spanish", URL: "/spanish", Note: spanishIndexNote(total)},
	}
}

// readingNote is the reading line's count: books open now and finished all-time.
func readingNote(reads []repo.ReadEntry) string {
	open := len(currentReads(reads))
	read := countStatus(reads, "finished")
	if open > 0 {
		return fmt.Sprintf("%d open; %d read", open, read)
	}
	return fmt.Sprintf("%d read", read)
}

// spanishIndexNote is the Spanish line's count: total hours of input logged.
func spanishIndexNote(total int) string {
	if total <= 0 {
		return "not started"
	}
	return fmt.Sprintf("%dh logged", total/3600)
}

// tagStrip matches HTML tags for reducing rendered markup to plain text.
var tagStrip = regexp.MustCompile(`<[^>]*>`)

// teaser reduces rendered note HTML to a plain-text lead, truncated at max on a
// word boundary with an ellipsis.
func teaser(h template.HTML, max int) string {
	s := html.UnescapeString(tagStrip.ReplaceAllString(string(h), " "))
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= max {
		return s
	}
	s = s[:max]
	if i := strings.LastIndex(s, " "); i > 0 {
		s = s[:i]
	}
	return s + "…"
}
