package site

import (
	"strconv"

	"github.com/henryppercy/hp-source/internal/site/templates"
)

const (
	everythingHeading    = "Everything"
	everythingSubheading = "Posts, slices, completed reads and spanish milestones from every feed in one stream, in the reverse order that they came in."
)

// everythingFeed assembles the /everything page from the complete merged
// stream.
func everythingFeed(stream []templates.FeedEntry) templates.EverythingFeedView {
	var posts, notes, books, milestones int
	for _, e := range stream {
		switch e.Kind {
		case "post":
			posts++
		case "note":
			notes++
		case "book":
			books++
		case "milestone":
			milestones++
		}
	}

	figures := []templates.Stat{{Label: "entries", Value: strconv.Itoa(len(stream))}}
	for _, f := range []templates.Stat{
		{Label: "posts", Value: strconv.Itoa(posts)},
		{Label: "notes", Value: strconv.Itoa(notes)},
		{Label: "books", Value: strconv.Itoa(books)},
		{Label: "milestones", Value: strconv.Itoa(milestones)},
	} {
		if f.Value != "0" {
			figures = append(figures, f)
		}
	}

	return templates.EverythingFeedView{
		Heading:    everythingHeading,
		Subheading: everythingSubheading,
		Empty:      feedEmpty,
		Figures:    figures,
		Entries:    stream,
	}
}
