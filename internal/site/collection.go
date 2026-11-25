package site

import (
	"fmt"

	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/site/templates"
	"github.com/henryppercy/hp-source/internal/text"
)

// authorURLs maps author names to their page URLs, filled by loadAuthorPages
// before any link is built. A name missing from the map has no page.
var authorURLs map[string]string

// authorURL returns the author's page URL, "" when they have no page.
func authorURL(name string) string {
	return authorURLs[name]
}

// authorPage is one author's collection page, the name with its slug and
// books.
type authorPage struct {
	Name  string
	Slug  string
	Books []repo.CollectionBook
}

// loadAuthorPages fetches every author with a paged book and fills
// authorURLs. It errors when two authors share a slug, since one would
// silently overwrite the other's page.
func (b *builder) loadAuthorPages() ([]authorPage, error) {
	authors, err := b.repo.ListAuthors()
	if err != nil {
		return nil, err
	}
	authorURLs = make(map[string]string, len(authors))
	bySlug := make(map[string]int)
	var pages []authorPage
	for _, a := range authors {
		books, err := b.repo.ListBooksByAuthor(a.ID)
		if err != nil {
			return nil, err
		}
		if len(books) == 0 {
			continue
		}
		slug := text.Slug(a.Name)
		if slug == "" {
			slug = fmt.Sprintf("author-%d", a.ID)
		}
		if other, ok := bySlug[slug]; ok {
			return nil, fmt.Errorf("authors %d and %d share slug %q, give one a distinct name", other, a.ID, slug)
		}
		bySlug[slug] = a.ID
		authorURLs[a.Name] = "/reading/author/" + slug
		pages = append(pages, authorPage{Name: a.Name, Slug: slug, Books: books})
	}
	return pages, nil
}

// renderAuthorPages writes each author's page under /reading/author so the
// reading nav stays the active section.
func (b *builder) renderAuthorPages(pages []authorPage) error {
	for _, p := range pages {
		view := collectionView("Author", p.Name, p.Books)
		if err := b.render("/reading/author/"+p.Slug, templates.Collection(view)); err != nil {
			return err
		}
	}
	return nil
}

// collectionView assembles the shared view for a collection page, the kind
// label over the name, the tally strip under it, and the books split into
// status sections in the repo's publication order.
func collectionView(kind, name string, books []repo.CollectionBook) templates.CollectionView {
	var sections [collectionSectionCount]templates.CollectionSection
	for i := range sections {
		sections[i].Label = collectionSectionLabels[i]
	}
	for _, e := range books {
		b := collectionBook(e)
		i := collectionSectionIndex(e.Status)
		sections[i].Books = append(sections[i].Books, b)
	}
	v := templates.CollectionView{
		Kicker:  kind,
		Title:   name,
		Figures: collectionFigures(books),
	}
	for _, s := range sections {
		if len(s.Books) > 0 {
			v.Sections = append(v.Sections, s)
		}
	}
	return v
}

// collectionSectionLabels names the status sections in page order, open
// books leading, the unread trailing, matching the hub's desk-to-shelf run.
var collectionSectionLabels = [collectionSectionCount]string{
	"Reading", "Read", "Set aside", "Unread",
}

// collectionSectionCount caps the status sections.
const collectionSectionCount = 4

// collectionSectionIndex maps a book status to its section slot, unknown
// statuses landing with the unread.
func collectionSectionIndex(status string) int {
	switch status {
	case "reading":
		return 0
	case "finished":
		return 1
	case "abandoned":
		return 2
	default:
		return 3
	}
}

// collectionBook turns one repo entry into a grid cell, the read's dates
// folded into the fields the meta line prints.
func collectionBook(e repo.CollectionBook) templates.CollectionBookView {
	v := templates.CollectionBookView{
		URL:        bookURL(e.BookID),
		Title:      e.Title,
		ImageURL:   coverURL(e.CoverImage),
		Status:     e.Status,
		Rating:     e.Rating,
		Percent:    readingPercent(e.CurrentPage, e.PageCount),
		FinishedAt: parseDate(e.DateFinished),
		AcquiredAt: parseDate(e.DateAcquired),
		Pages:      e.PageCount,
	}
	if published := parseDate(e.DatePublished); !published.IsZero() {
		v.Published = published.Format("2006")
	}
	if e.Status == "reading" {
		v.DayCount = daysSince(parseDate(e.DateStarted))
	}
	return v
}

// collectionFigures is the tally strip under a collection's title, the total
// and the read count always shown, the other statuses only when they hold
// books so the strip never counts an empty state.
func collectionFigures(books []repo.CollectionBook) []templates.Stat {
	read, reading, aside, unread := 0, 0, 0, 0
	for _, e := range books {
		switch e.Status {
		case "finished":
			read++
		case "reading":
			reading++
		case "abandoned":
			aside++
		default:
			unread++
		}
	}
	figures := []templates.Stat{
		{Label: "books", Value: fmt.Sprintf("%d", len(books))},
		{Label: "read", Value: fmt.Sprintf("%d", read)},
	}
	if reading > 0 {
		figures = append(figures, templates.Stat{Label: "reading", Value: fmt.Sprintf("%d", reading)})
	}
	if aside > 0 {
		figures = append(figures, templates.Stat{Label: "set aside", Value: fmt.Sprintf("%d", aside)})
	}
	if unread > 0 {
		figures = append(figures, templates.Stat{Label: "unread", Value: fmt.Sprintf("%d", unread)})
	}
	return figures
}
