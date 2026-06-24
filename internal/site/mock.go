package site

import "time"

// Renders from hardcoded mock view-models. These are replaced by repo
// data in a later stage; the view-model shapes stay the same.

func date(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func mockPosts() []PostView {
	return []PostView{
		{
			Title:     "Hello World",
			Slug:      "hello-world",
			Type:      "",
			PostedAt:  date("2026-01-10"),
			UpdatedAt: date("2026-02-01"),
			Headline:  "The first post on the new generator.",
			BodyHTML:  "<p>This is a placeholder body. Markdown rendering arrives in stage 2.</p>",
		},
		{
			Title:    "A Short Note",
			Slug:     "a-short-note",
			Type:     "slice",
			PostedAt: date("2026-03-04"),
			Headline: "A quick timestamped thought.",
			BodyHTML: "<p>Slices are short, timestamped notes.</p>",
		},
		{
			Title:    "Reached B1",
			Slug:     "reached-b1",
			Type:     "spanish",
			PostedAt: date("2026-05-20"),
			Headline: "A Spanish-learning milestone.",
			BodyHTML: "<p>Level B1 reached. Details live in the body.</p>",
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

func toListItem(p PostView) PostListItem {
	return PostListItem{
		Title:    p.Title,
		Slug:     p.Slug,
		Type:     p.Type,
		PostedAt: p.PostedAt,
		Headline: p.Headline,
	}
}

func toListItems(posts []PostView) []PostListItem {
	var items []PostListItem
	for _, p := range posts {
		items = append(items, toListItem(p))
	}
	return items
}

func postListItems(posts []PostView, typ string) []PostListItem {
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
