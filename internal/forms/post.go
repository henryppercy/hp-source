package forms

import (
	"fmt"
	"strings"
	"time"

	huh "charm.land/huh/v2"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/text"
)

func AddPost(in *repo.PostInput) error  { return postForm(in, "Add Post", "Create") }
func EditPost(in *repo.PostInput) error { return postForm(in, "Edit Post", "Save") }

// postForm collects an article's metadata (post or spanish). Slices use
// sliceForm. The slug is finalised by the repo (override, else title), so the
// form just collects the raw optional override.
func postForm(in *repo.PostInput, heading, saveLabel string) error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(heading).
				Next(true).
				NextLabel("Next"),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Placeholder("required").
				Validate(huh.ValidateNotEmpty()).
				Value(&in.Title),
			huh.NewInput().
				Title("Headline").
				Placeholder("optional").
				Value(&in.Headline),
			huh.NewSelect[string]().
				Title("Type").
				Options(
					huh.NewOption("Post", ""),
					huh.NewOption("Spanish", "spanish"),
				).
				Value(&in.Type),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Slug").
				PlaceholderFunc(func() string {
					return text.Slug(in.Title)
				}, &in.Title).
				Value(&in.Slug),
			huh.NewInput().
				Title("Published At").
				Placeholder("blank = draft, e.g. 2026-06-20").
				Validate(validateDateOptional).
				Value(&in.PublishedAt),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Confirm").
				DescriptionFunc(func() string {
					return postSummary(in)
				}, &in.PublishedAt).
				Next(true).
				NextLabel(saveLabel),
		),
	)

	return form.Run()
}

func AddSlice(in *repo.PostInput) error  { return sliceForm(in, "Add Slice") }
func EditSlice(in *repo.PostInput) error { return sliceForm(in, "Edit Slice") }

// sliceForm collects a slice's thin metadata: an optional slug override (else a
// date slug is assigned by the repo) and the publish date.
func sliceForm(in *repo.PostInput, heading string) error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Slug").
				Description(heading).
				Placeholder("auto from date").
				Value(&in.Slug),
			huh.NewInput().
				Title("Published At").
				Placeholder("blank = draft, e.g. 2026-06-20").
				Validate(validateDateOptional).
				Value(&in.PublishedAt),
		),
	)
	return form.Run()
}

// SelectArticle asks which article to act on, labelled by title. Drafts marked.
func SelectArticle(posts []repo.Post) (int, error) {
	return selectPost(posts, func(p repo.Post) string {
		label := p.Title
		if p.PublishedAt == "" {
			label += " (draft)"
		}
		return label
	})
}

// SelectSlice asks which slice to act on, labelled by date + a body preview
// (slices have no title). Drafts are marked.
func SelectSlice(slices []repo.Post) (int, error) {
	return selectPost(slices, func(p repo.Post) string {
		date := p.PublishedAt
		if date == "" {
			date = p.CreatedAt
		}
		label := fmt.Sprintf("%s  %s", sliceDate(date), preview(p.Body))
		if p.PublishedAt == "" {
			label += " (draft)"
		}
		return label
	})
}

func selectPost(posts []repo.Post, label func(repo.Post) string) (int, error) {
	options := make([]huh.Option[int], len(posts))
	for i, p := range posts {
		options[i] = huh.NewOption(label(p), p.ID)
	}

	var id int
	sel := huh.NewSelect[int]().
		Title("Which one?").
		Options(options...).
		Value(&id)
	if len(posts) > 10 {
		sel.Height(10)
	}

	if err := huh.NewForm(huh.NewGroup(sel)).Run(); err != nil {
		return 0, err
	}
	return id, nil
}

// sliceDate trims a stored date/datetime to its date part for labels.
func sliceDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// preview is a slice body's first non-empty line, trimmed to ~50 chars.
func preview(body string) string {
	line := ""
	for _, l := range strings.Split(body, "\n") {
		if strings.TrimSpace(l) != "" {
			line = strings.TrimSpace(l)
			break
		}
	}
	if len(line) > 50 {
		line = strings.TrimSpace(line[:50]) + "..."
	}
	return line
}

// PublishPost prompts for the publish date, defaulting an empty answer to today.
func PublishPost(title string, date *string) error {
	today := time.Now().Format("2006-01-02")
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Published At").
				Description("Publishing: " + title).
				Placeholder(today).
				Validate(validateDateOptional).
				Value(date),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}
	if *date == "" {
		*date = today
	}
	return nil
}

func validateDateOptional(s string) error {
	if s == "" {
		return nil
	}
	return validateDate(s)
}

func postSummary(in *repo.PostInput) string {
	var sb strings.Builder

	typeLabel := "Post"
	switch in.Type {
	case "slice":
		typeLabel = "Slice"
	case "spanish":
		typeLabel = "Spanish"
	}
	fmt.Fprintf(&sb, "Type:       %s\n", typeLabel)

	if in.Title != "" {
		fmt.Fprintf(&sb, "Title:      %s\n", in.Title)
	}

	slug := text.Slug(in.Slug)
	if slug == "" {
		slug = text.Slug(in.Title)
	}
	if slug == "" {
		slug = "auto (date)"
	}
	fmt.Fprintf(&sb, "Slug:       %s\n", slug)

	if in.Headline != "" {
		fmt.Fprintf(&sb, "Headline:   %s\n", in.Headline)
	}

	published := in.PublishedAt
	if published == "" {
		published = "draft"
	}
	fmt.Fprintf(&sb, "Published:  %s\n", published)

	return sb.String()
}
