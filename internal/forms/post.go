package forms

import (
	"fmt"
	"strings"
	"time"

	huh "charm.land/huh/v2"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/text"
)

// AddPost collects a new article's details into in.
func AddPost(in *repo.PostInput, topics []repo.Topic, locations []repo.Location) error {
	return postForm(in, topics, locations, "Add Post", "Create")
}

// EditPost collects changes to an article's details into in.
func EditPost(in *repo.PostInput, topics []repo.Topic, locations []repo.Location) error {
	return postForm(in, topics, locations, "Edit Post", "Save")
}

func postForm(in *repo.PostInput, topics []repo.Topic, locations []repo.Location, heading, saveLabel string) error {
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
				Placeholder("blank = draft, e.g. 2026-06-20 or 2026-06-20 14:32").
				Validate(validateDateTimeOptional).
				Value(&in.PublishedAt),
		),

		locationsGroup(in, locations),

		topicsGroup(in, topics),

		huh.NewGroup(
			huh.NewNote().
				Title("Confirm").
				DescriptionFunc(func() string {
					return postSummary(in, topics, locations)
				}, &in.PublishedAt).
				Next(true).
				NextLabel(saveLabel),
		),
	)

	return form.Run()
}

// AddSlice collects a new slice's details into in.
func AddSlice(in *repo.PostInput, topics []repo.Topic, locations []repo.Location) error {
	return sliceForm(in, topics, locations, "Add Slice")
}

// EditSlice collects changes to a slice's details into in.
func EditSlice(in *repo.PostInput, topics []repo.Topic, locations []repo.Location) error {
	return sliceForm(in, topics, locations, "Edit Slice")
}

func sliceForm(in *repo.PostInput, topics []repo.Topic, locations []repo.Location, heading string) error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Slug").
				Description(heading).
				Placeholder("auto from date").
				Value(&in.Slug),
			huh.NewInput().
				Title("Published At").
				Placeholder("blank = draft, e.g. 2026-06-20 or 2026-06-20 14:32").
				Validate(validateDateTimeOptional).
				Value(&in.PublishedAt),
		),

		locationsGroup(in, locations),

		topicsGroup(in, topics),
	)
	return form.Run()
}

func locationsGroup(in *repo.PostInput, locations []repo.Location) *huh.Group {
	options := []huh.Option[int]{huh.NewOption("None", 0)}
	for _, l := range locations {
		options = append(options, huh.NewOption(l.Name, l.ID))
	}
	sel := huh.NewSelect[int]().
		Title("Location").
		Options(options...).
		Value(&in.LocationID)
	if len(options) > 10 {
		sel.Height(10)
	}
	return huh.NewGroup(sel)
}

func topicsGroup(in *repo.PostInput, topics []repo.Topic) *huh.Group {
	options := make([]huh.Option[int], len(topics))
	for i, t := range topics {
		options[i] = huh.NewOption(t.Name, t.ID)
	}
	return huh.NewGroup(
		huh.NewMultiSelect[int]().
			Title("Topics").
			Options(options...).
			Value(&in.TopicIDs),
	)
}

// SelectArticle picks an article from posts, labelling drafts, and returns its
// id.
func SelectArticle(posts []repo.Post) (int, error) {
	return selectPost(posts, func(p repo.Post) string {
		label := p.Title
		if p.PublishedAt == "" {
			label += " (draft)"
		}
		return label
	})
}

// SelectSlice picks a slice from slices and returns its id.
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

// PublishPost collects the published date for title into date, defaulting to
// now.
func PublishPost(title string, date *string) error {
	now := time.Now()
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Published At").
				Description("Publishing: " + title).
				Placeholder(now.Format("2006-01-02")).
				Validate(validateDateTimeOptional).
				Value(date),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}
	if *date == "" {
		*date = now.Format("2006-01-02 15:04:05")
	}
	return nil
}

func postSummary(in *repo.PostInput, topics []repo.Topic, locations []repo.Location) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Title:      %s\n", in.Title)

	slug := text.Slug(in.Slug)
	if slug == "" {
		slug = text.Slug(in.Title)
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

	for _, l := range locations {
		if l.ID == in.LocationID {
			fmt.Fprintf(&sb, "Location:   %s\n", l.Name)
			break
		}
	}

	if names := topicNames(in.TopicIDs, topics); len(names) > 0 {
		fmt.Fprintf(&sb, "Topics:     %s\n", strings.Join(names, ", "))
	}

	return sb.String()
}

func topicNames(ids []int, topics []repo.Topic) []string {
	var names []string
	for _, id := range ids {
		for _, t := range topics {
			if t.ID == id {
				names = append(names, t.Name)
				break
			}
		}
	}
	return names
}
