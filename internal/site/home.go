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

// homeCurrently is the free-text "what I'm up to" line in the dispatch strip.
const homeCurrently = "Peeling myself of the ceiling after England v Mexico, " +
	"building Ikea flatpack, and pulling yellow eletrical tape out of my laptop."

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
	milestones := spanishMilestones(days, total, articleURLs(articlesWithTopic(posts, "spanish")))

	return templates.HomeView{
		Copy:     homeCopy,
		Dispatch: dispatchCells(reads, days, total, articles, notes, now),
		Stats:    colophonStats(reads, articles, notes, total),
		Subjects: topicCounts(posts),
		Stream:   lifeStream(articles, notes, reads, milestoneEntries(milestones)),
		Index:    indexRows(articles, notes, reads, days, now),
	}
}

// topicCounts ranks the subjects across all posts by how many carry each, most
// first. Spanish is left out; it has its own section and index line.
func topicCounts(posts []repo.Post) []templates.TopicCount {
	counts := map[string]int{}
	var order []string
	for _, p := range posts {
		for _, t := range p.Topics {
			if t.Name == "spanish" {
				continue
			}
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
			meta += fmt.Sprintf(" ; day %d", r.DayCount)
		}
		cells = append(cells, templates.DispatchCell{
			Kicker: "Open on the desk",
			Lead:   r.Title + ", " + r.Author,
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
			meta += fmt.Sprintf(" ; %dd streak", cur)
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

// colophonStats are the figures summing up the whole notebook.
func colophonStats(
	reads []repo.ReadEntry,
	articles []templates.PostListItem,
	notes []templates.SliceItem,
	total int,
) []templates.Stat {
	return []templates.Stat{
		{Label: "books read", Value: fmt.Sprintf("%d", countStatus(reads, "finished"))},
		{Label: "spanish hours logged", Value: fmt.Sprintf("%d", total/3600)},
		{Label: "articles filed", Value: fmt.Sprintf("%d", len(articles))},
		{Label: "notes made", Value: fmt.Sprintf("%d", len(notes))},
	}
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
	return strings.Join(parts, " ; ")
}

// milestoneEntries turns the reached roadmap rungs into stream entries, dropping
// the day-one zero rung, which marks a start rather than a crossing.
func milestoneEntries(rungs []templates.MilestoneRung) []templates.FeedEntry {
	var out []templates.FeedEntry
	for _, r := range rungs {
		if !r.Reached || r.Date.IsZero() || strings.HasPrefix(r.Label, "0 ") {
			continue
		}
		out = append(out, templates.FeedEntry{
			Kind: "milestone", Kicker: "Español", Date: r.Date, Title: r.Label, URL: r.URL,
		})
	}
	return out
}

// indexRows build the section directory with a live count against each line.
func indexRows(
	articles []templates.PostListItem,
	notes []templates.SliceItem,
	reads []repo.ReadEntry,
	days []spanishDay,
	now time.Time,
) []templates.IndexRow {
	return []templates.IndexRow{
		{Num: "02", Label: "Posts", URL: "/posts", Note: fmt.Sprintf("%d filed", len(articles))},
		{Num: "03", Label: "Slices", URL: "/slices", Note: fmt.Sprintf("%d notes", len(notes))},
		{Num: "04", Label: "Reading", URL: "/reading", Note: readingNote(reads, now.Year())},
		{Num: "05", Label: "Spanish", URL: "/spanish", Note: spanishIndexNote(days, now)},
	}
}

// readingNote is the reading line's count: books open now and finished this year.
func readingNote(reads []repo.ReadEntry, year int) string {
	open := len(currentReads(reads))
	read := booksReadInYear(reads, year)
	if open > 0 {
		return fmt.Sprintf("%d open ; %d this year", open, read)
	}
	return fmt.Sprintf("%d this year", read)
}

// spanishIndexNote is the Spanish line's count: the current day of input.
func spanishIndexNote(days []spanishDay, now time.Time) string {
	if len(days) == 0 {
		return "not started"
	}
	return fmt.Sprintf("day %d", spanishDayCount(days, now))
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
