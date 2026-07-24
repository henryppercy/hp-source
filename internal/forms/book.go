package forms

import (
	"fmt"
	"strconv"
	"strings"

	huh "charm.land/huh/v2"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/text"
)

// AddBook collects a work's details, returning whether the user wants to add a
// copy. Authors, series and tags are selected from existing records; create them
// first with author/series/tag add.
func AddBook(
	input *repo.BookInput,
	genres []repo.Genre,
	authors []repo.Author,
	tags []repo.Tag,
	seriesList []repo.Series,
) (bool, error) {
	genreOptions := make([]huh.Option[int], len(genres))
	for i, g := range genres {
		genreOptions[i] = huh.NewOption(g.Name, g.ID)
	}

	authorOptions := make([]huh.Option[int], len(authors))
	for i, a := range authors {
		authorOptions[i] = huh.NewOption(a.Name, a.ID)
	}

	tagOptions := make([]huh.Option[int], len(tags))
	for i, t := range tags {
		tagOptions[i] = huh.NewOption(t.Name, t.ID)
	}

	seriesOptions := make([]huh.Option[int], len(seriesList))
	for i, s := range seriesList {
		seriesOptions[i] = huh.NewOption(s.Name, s.ID)
	}

	var (
		selectedAuthorIDs []int
		selectedSeriesID  int
		seriesPosition    string
		addSeries         bool
		addTags           bool
		addCopy           bool
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Add Book").
				Description("Add a new work to your collection.").
				Next(true).
				NextLabel("Next"),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Placeholder("required").
				Validate(huh.ValidateNotEmpty()).
				Value(&input.Title),
			huh.NewSelect[string]().
				Title("Type").
				Options(huh.NewOptions("fiction", "non-fiction")...).
				Value(&input.BookType),
			huh.NewSelect[int]().
				Title("Genre").
				Height(15).
				Options(genreOptions...).
				Value(&input.GenreID),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Date Published").
				Placeholder("required, e.g. 1963 or 1963-11-22").
				Validate(validateDateOrYear).
				Value(&input.DatePublished),
			huh.NewInput().
				Title("Original Language").
				Placeholder("english").
				Value(&input.OriginalLanguage),
			huh.NewInput().
				Title("URL").
				Placeholder("optional").
				Value(&input.URL),
		),

		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Authors").
				Height(15).
				Options(authorOptions...).
				Validate(func(ids []int) error {
					if len(ids) == 0 {
						return fmt.Errorf("select at least one author")
					}
					return nil
				}).
				Value(&selectedAuthorIDs),
		),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Part of a series?").
				Affirmative("Yes").
				Negative("No").
				Value(&addSeries),
		),

		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Series").
				Options(seriesOptions...).
				Value(&selectedSeriesID),
			huh.NewInput().
				Title("Series Position").
				Placeholder("e.g. 1, 1.5").
				Validate(huh.ValidateNotEmpty()).
				Value(&seriesPosition),
		).WithHideFunc(func() bool {
			return !addSeries
		}),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Add tags?").
				Affirmative("Yes").
				Negative("No").
				Value(&addTags),
		),

		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Tags").
				Height(15).
				Limit(6).
				Options(tagOptions...).
				Value(&input.TagIDs),
		).WithHideFunc(func() bool {
			return !addTags
		}),

		huh.NewGroup(
			huh.NewNote().
				Title("Confirm Book Details").
				DescriptionFunc(func() string {
					return bookSummary(input, genres, authors, seriesList, tags,
						selectedAuthorIDs, addSeries, selectedSeriesID, seriesPosition)
				}, &input.Title).
				Next(true).
				NextLabel("Next"),
		),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Add a copy now?").
				Description("A copy is an edition you own or want.").
				Affirmative("Yes").
				Negative("No").
				Value(&addCopy),
		),
	)

	if err := form.Run(); err != nil {
		return false, err
	}

	if input.OriginalLanguage == "" {
		input.OriginalLanguage = "english"
	}

	for _, id := range selectedAuthorIDs {
		for _, a := range authors {
			if a.ID == id {
				input.Authors = append(input.Authors, repo.AuthorInput{
					ID:   a.ID,
					Name: a.Name,
					Role: "author",
				})
				break
			}
		}
	}

	if addSeries && selectedSeriesID != 0 {
		pos, _ := strconv.ParseFloat(seriesPosition, 64)
		input.Series = &repo.SeriesInput{
			ID:       selectedSeriesID,
			Position: pos,
		}
	}

	return addCopy, nil
}

func bookSummary(
	input *repo.BookInput,
	genres []repo.Genre,
	authors []repo.Author,
	seriesList []repo.Series,
	tags []repo.Tag,
	selectedAuthorIDs []int,
	addSeries bool,
	selectedSeriesID int,
	seriesPosition string,
) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "Title:      %s\n", input.Title)
	fmt.Fprintf(&sb, "Type:       %s\n", input.BookType)

	for _, g := range genres {
		if g.ID == input.GenreID {
			fmt.Fprintf(&sb, "Genre:      %s\n", g.Name)
			break
		}
	}

	fmt.Fprintf(&sb, "Published:  %s\n", input.DatePublished)

	lang := input.OriginalLanguage
	if lang == "" {
		lang = "english"
	}
	fmt.Fprintf(&sb, "Language:   %s\n", lang)

	if input.URL != "" {
		fmt.Fprintf(&sb, "URL:        %s\n", input.URL)
	}

	fmt.Fprintf(&sb, "\n")
	for _, id := range selectedAuthorIDs {
		for _, a := range authors {
			if a.ID == id {
				fmt.Fprintf(&sb, "Author:     %s\n", a.Name)
				break
			}
		}
	}

	if addSeries {
		for _, s := range seriesList {
			if s.ID == selectedSeriesID {
				fmt.Fprintf(&sb, "\nSeries:     %s (#%s)\n", s.Name, seriesPosition)
				break
			}
		}
	}

	if len(input.TagIDs) > 0 {
		fmt.Fprintf(&sb, "\n")
		var tagNames []string
		for _, id := range input.TagIDs {
			for _, t := range tags {
				if t.ID == id {
					tagNames = append(tagNames, t.Name)
					break
				}
			}
		}
		fmt.Fprintf(&sb, "Tags:       %s\n", strings.Join(tagNames, ", "))
	}

	return sb.String()
}

// EditBook edits a work's fields, prefilled from its current values.
func EditBook(in *repo.BookEdit, genres []repo.Genre) error {
	genreOptions := make([]huh.Option[int], len(genres))
	for i, g := range genres {
		genreOptions[i] = huh.NewOption(g.Name, g.ID)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Edit Book").
				Next(true).
				NextLabel("Next"),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Validate(huh.ValidateNotEmpty()).
				Value(&in.Title),
			huh.NewSelect[string]().
				Title("Type").
				Options(huh.NewOptions("fiction", "non-fiction")...).
				Value(&in.BookType),
			huh.NewSelect[int]().
				Title("Genre").
				Height(15).
				Options(genreOptions...).
				Value(&in.GenreID),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Date Published").
				Validate(validateDateOrYear).
				Value(&in.DatePublished),
			huh.NewInput().
				Title("Original Language").
				Validate(huh.ValidateNotEmpty()).
				Value(&in.OriginalLanguage),
			huh.NewInput().
				Title("URL").
				Placeholder("optional").
				Value(&in.URL),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Confirm Book").
				DescriptionFunc(func() string {
					var sb strings.Builder
					fmt.Fprintf(&sb, "Title:      %s\n", in.Title)
					fmt.Fprintf(&sb, "Type:       %s\n", in.BookType)
					for _, g := range genres {
						if g.ID == in.GenreID {
							fmt.Fprintf(&sb, "Genre:      %s\n", g.Name)
							break
						}
					}
					fmt.Fprintf(&sb, "Published:  %s\n", in.DatePublished)
					fmt.Fprintf(&sb, "Language:   %s\n", in.OriginalLanguage)
					if in.URL != "" {
						fmt.Fprintf(&sb, "URL:        %s\n", in.URL)
					}
					return sb.String()
				}, &in.Title).
				Next(true).
				NextLabel("Save"),
		),
	)

	return form.Run()
}

func sortName(name string) string {
	parts := strings.Fields(name)
	if len(parts) >= 2 {
		last := parts[len(parts)-1]
		rest := strings.Join(parts[:len(parts)-1], " ")
		return last + ", " + rest
	}
	return name
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func coverImageName(title string) string {
	if title == "" {
		return ""
	}
	return text.Slug(title) + ".jpg"
}
