package site

import "time"

// Renders from hardcoded mock view-models. These are replaced by repo
// data in a later stage; the view-model shapes stay the same.

// postSource is mock stand-in for a stored post: metadata plus the raw
// markdown body. The builder renders the body into a PostView.
type postSource struct {
	Title     string
	Slug      string
	Type      string
	PostedAt  time.Time
	UpdatedAt time.Time
	Headline  string
	Markdown  string
}

func date(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func mockPosts() []postSource {
	return []postSource{
		{
			Title:     "Hello World",
			Slug:      "hello-world",
			Type:      "",
			PostedAt:  date("2026-01-10"),
			UpdatedAt: date("2026-02-01"),
			Headline:  "The first post on the new generator.",
			Markdown:  fatPostMarkdown,
		},
		{
			Title:    "A Short Note",
			Slug:     "a-short-note",
			Type:     "slice",
			PostedAt: date("2026-03-04"),
			Headline: "A quick timestamped thought.",
			Markdown: "Slices are short, timestamped notes. Nothing fancy here.",
		},
		{
			Title:    "Reached B1",
			Slug:     "reached-b1",
			Type:     "spanish",
			PostedAt: date("2026-05-20"),
			Headline: "A Spanish-learning milestone.",
			Markdown: "**Level:** B1\n\n**Date achieved:** 2026-05-20\n\nReached B1. Details live in the body.",
		},
	}
}

func mockBooks() []BookView {
	return []BookView{
		{
			Title:        "Currently Reading This",
			Author:       "Some Author",
			ImageURL:     "/static/placeholder-cover.png",
			Status:       "reading",
			Rating:       "",
			DateFinished: time.Time{},
		},
		{
			Title:        "A Finished Book",
			Author:       "Another Author",
			ImageURL:     "/static/placeholder-cover.png",
			Status:       "finished",
			Rating:       "4.5",
			DateFinished: date("2026-04-12"),
		},
	}
}

func toListItem(p postSource) PostListItem {
	return PostListItem{
		Title:    p.Title,
		Slug:     p.Slug,
		Type:     p.Type,
		PostedAt: p.PostedAt,
		Headline: p.Headline,
	}
}

func toListItems(posts []postSource) []PostListItem {
	var items []PostListItem
	for _, p := range posts {
		items = append(items, toListItem(p))
	}
	return items
}

func postListItems(posts []postSource, typ string) []PostListItem {
	var items []PostListItem
	for _, p := range posts {
		if p.Type == typ {
			items = append(items, toListItem(p))
		}
	}
	return items
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

// fatPostMarkdown deliberately exercises many heading levels, code blocks in
// several languages, and assorted inline markdown to stress the renderer/TOC.
const fatPostMarkdown = "" +
	"This intro paragraph sits before any heading. It has **bold**, *italic*, " +
	"`inline code` and a [link](https://example.com).\n\n" +
	"## Getting started\n\n" +
	"Some prose under the first section.\n\n" +
	"### Installation\n\n" +
	"A nested subsection with a shell block:\n\n" +
	"```sh\ngo install github.com/henryppercy/hp@latest\nhp site build\n```\n\n" +
	"### Configuration\n\n" +
	"Another subsection, this time with Go:\n\n" +
	"```go\nfunc main() {\n\tfmt.Println(\"hello, site\")\n}\n```\n\n" +
	"## Usage\n\n" +
	"A second top-level section.\n\n" +
	"- first item\n- second item\n- third item\n\n" +
	"### Examples\n\n" +
	"Some JSON:\n\n" +
	"```json\n{\n  \"name\": \"hp\",\n  \"stage\": 2\n}\n```\n\n" +
	"> A blockquote to round things out.\n\n" +
	"## Wrapping up\n\n" +
	"The final section, no subsections.\n"
