package site

import (
	"fmt"
	"math"
	"strings"

	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/site/templates"
	"github.com/henryppercy/hp-source/internal/text"
)

// bookUnread heads the note shown on a page for a book never started.
const bookUnread = "Not read yet."

// bookURLs maps book ids to page URLs, filled by loadBookPages before any
// link is built. An id missing from the map has no page, so bookURL falls
// back to /reading.
var bookURLs map[int]string

func bookURL(bookID int) string {
	if u, ok := bookURLs[bookID]; ok {
		return u
	}
	return "/reading"
}

// loadBookPages fetches every book with a copy or a read and fills bookURLs.
// It errors when two books share a slug, since one would silently overwrite
// the other's page.
func (b *builder) loadBookPages() ([]*repo.BookPage, error) {
	ids, err := b.repo.ListBookPageIDs()
	if err != nil {
		return nil, err
	}

	bookURLs = make(map[int]string, len(ids))
	bySlug := make(map[string]int, len(ids))
	pages := make([]*repo.BookPage, 0, len(ids))
	for _, id := range ids {
		p, err := b.repo.GetBookPage(id)
		if err != nil {
			return nil, err
		}
		slug := bookSlug(p)
		if other, ok := bySlug[slug]; ok {
			return nil, fmt.Errorf("books %d and %d share slug %q, give one a distinct title", other, id, slug)
		}
		bySlug[slug] = id
		bookURLs[id] = "/reading/book/" + slug
		pages = append(pages, p)
	}
	return pages, nil
}

// renderBookPages writes every book page under /reading/book so the reading
// nav stays the active section.
func (b *builder) renderBookPages(pages []*repo.BookPage) error {
	for _, p := range pages {
		url := bookURLs[p.ID]
		if err := b.render(url, templates.Book(bookView(p))); err != nil {
			return err
		}
	}
	return nil
}

// bookSlug falls back to the id when a title slugs to nothing, as a
// non-latin one does, so every page has a stable URL.
func bookSlug(p *repo.BookPage) string {
	if slug := text.Slug(p.Title); slug != "" {
		return slug
	}
	return fmt.Sprintf("book-%d", p.ID)
}

func bookView(p *repo.BookPage) templates.BookView {
	tags := make([]string, len(p.Tags))
	for i, t := range p.Tags {
		tags[i] = text.TitleCase(t)
	}
	reads := bookReads(p)
	v := templates.BookView{
		Title:       p.Title,
		Kind:        text.TitleCase(p.BookType),
		ImageURL:    coverURL(p.CoverImage),
		Authors:     bookAuthors(p.Authors),
		Series:      seriesLabel(p.Series, p.SeriesPosition),
		Translators: bookTranslators(p.Translators),
		Tags:        tags,
		Facts:       bookFacts(p),
		Head:        bookReadsHead(p.Reads),
		Reads:       reads,
	}

	if p.OriginalTitle != "" && p.OriginalTitle != p.Title {
		v.OriginalTitle = p.OriginalTitle
	}
	if len(p.Reads) == 0 {
		v.UnreadNote = bookUnread + bookAcquired(p)
	}
	return v
}

func bookAuthors(credits []repo.BookCredit) string {
	var names, others []string
	for _, c := range credits {
		if c.Role == "author" {
			names = append(names, c.Name)
			continue
		}
		others = append(others, fmt.Sprintf("%s (%s)", c.Name, c.Role))
	}
	return strings.Join(append(names, others...), ", ")
}

func bookTranslators(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return "Translated by " + strings.Join(names, ", ")
}

// bookFacts lists the work's facts, each left out when unrecorded.
func bookFacts(p *repo.BookPage) []templates.Stat {
	facts := []templates.Stat{
		{Label: "genre", Value: text.TitleCase(p.Genre)},
	}
	if y := parseDate(p.DatePublished).Format("2006"); y != "0001" {
		facts = append(facts, templates.Stat{Label: "published", Value: y})
	}
	facts = append(facts, templates.Stat{Label: "language", Value: text.TitleCase(p.OriginalLanguage)})
	if p.PageCount > 0 {
		facts = append(facts, templates.Stat{Label: "pages", Value: fmt.Sprintf("%d", p.PageCount)})
	}
	return facts
}

func bookAcquired(p *repo.BookPage) string {
	acquired := parseDate(p.DateAcquired)
	if acquired.IsZero() {
		return ""
	}
	return fmt.Sprintf(" On the shelf since %s.", acquired.Format("2 Jan 2006"))
}

// bookReadsHead is the reads count for the section head, "" when the
// book has never been started.
func bookReadsHead(reads []repo.BookRead) string {
	switch len(reads) {
	case 0:
		return ""
	case 1:
		return "once"
	default:
		return fmt.Sprintf("%d times", len(reads))
	}
}

func bookReads(p *repo.BookPage) []templates.BookReadView {
	out := make([]templates.BookReadView, 0, len(p.Reads))
	for _, rd := range p.Reads {
		v := templates.BookReadView{
			Status:      rd.Status,
			Edition:     bookEdition(rd, p.Title),
			Rating:      rd.Rating,
			StartedAt:   parseDate(rd.DateStarted),
			FinishedAt:  parseDate(rd.DateFinished),
			Days:        daysBetween(parseDate(rd.DateStarted), parseDate(rd.DateFinished)),
			Pages:       rd.PageCount,
			Pace:        bookPace(rd),
			Format:      text.TitleCase(rd.Format),
			Language:    text.TitleCase(rd.Language),
			Source:      text.TitleCase(p.Source),
			SecondHand:  p.SecondHand,
			Percent:     readingPercent(rd.CurrentPage, rd.PageCount),
			CurrentPage: rd.CurrentPage,
		}
		if rd.Status == "reading" {
			v.DayCount = daysSince(v.StartedAt)
		}
		out = append(out, v)
	}
	return out
}

func bookEdition(rd repo.BookRead, page string) string {
	if rd.Title == "" || rd.Title == page {
		return ""
	}
	return rd.Title
}

func bookPace(rd repo.BookRead) string {
	days := daysBetween(parseDate(rd.DateStarted), parseDate(rd.DateFinished))
	if rd.Status != "finished" || days <= 0 || rd.PageCount <= 0 {
		return ""
	}
	return fmt.Sprintf("%d pages/day", int(math.Round(float64(rd.PageCount)/float64(days))))
}
