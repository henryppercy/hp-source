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
				Title("Slug").
				PlaceholderFunc(func() string {
					return text.Slug(in.Title)
				}, &in.Title).
				Value(&in.Slug),
			huh.NewSelect[string]().
				Title("Type").
				Options(
					huh.NewOption("Post", ""),
					huh.NewOption("Slice", "slice"),
					huh.NewOption("Spanish", "spanish"),
				).
				Value(&in.Type),
			huh.NewInput().
				Title("Headline").
				Placeholder("optional").
				Value(&in.Headline),
		),

		huh.NewGroup(
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

	if err := form.Run(); err != nil {
		return err
	}

	// Slug is always derived/normalised so it stays URL-safe and unique-friendly,
	// falling back to the title when left blank.
	slug := in.Slug
	if slug == "" {
		slug = in.Title
	}
	in.Slug = text.Slug(slug)
	return nil
}

// SelectPost asks which post to act on, returning its id. Drafts are marked.
func SelectPost(posts []repo.Post) (int, error) {
	options := make([]huh.Option[int], len(posts))
	for i, p := range posts {
		label := p.Title
		if p.PublishedAt == "" {
			label += " (draft)"
		}
		options[i] = huh.NewOption(label, p.ID)
	}

	var id int
	postSelect := huh.NewSelect[int]().
		Title("Which post?").
		Options(options...).
		Value(&id)
	if len(posts) > 10 {
		postSelect.Height(10)
	}

	if err := huh.NewForm(huh.NewGroup(postSelect)).Run(); err != nil {
		return 0, err
	}
	return id, nil
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

	fmt.Fprintf(&sb, "Title:      %s\n", in.Title)

	slug := in.Slug
	if slug == "" {
		slug = text.Slug(in.Title)
	}
	fmt.Fprintf(&sb, "Slug:       %s\n", slug)

	typeLabel := "Post"
	switch in.Type {
	case "slice":
		typeLabel = "Slice"
	case "spanish":
		typeLabel = "Spanish"
	}
	fmt.Fprintf(&sb, "Type:       %s\n", typeLabel)

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
