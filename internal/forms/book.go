package forms

import (
	"strconv"
	"strings"

	huh "charm.land/huh/v2"
	"github.com/henryppercy/hp-source/internal/repo"
)

func AddBook(
	input *repo.BookInput,
	genres []repo.Genre,
	authors []repo.Author,
	tags []repo.Tag,
	seriesList []repo.Series,
) error {
	genreOptions := make([]huh.Option[int], len(genres))
	for i, g := range genres {
		genreOptions[i] = huh.NewOption(g.Name, g.ID)
	}

	authorOptions := []huh.Option[int]{huh.NewOption("+ Add new author", 0)}
	for _, a := range authors {
		authorOptions = append(authorOptions, huh.NewOption(a.Name, a.ID))
	}

	tagOptions := make([]huh.Option[int], len(tags))
	for i, t := range tags {
		tagOptions[i] = huh.NewOption(t.Name, t.ID)
	}

	seriesOptions := []huh.Option[int]{huh.NewOption("+ Add new series", 0)}
	for _, s := range seriesList {
		seriesOptions = append(seriesOptions, huh.NewOption(s.Name, s.ID))
	}

	var (
		selectedAuthorID int
		authorName       string
		authorSortName   string
		selectedSeriesID int
		newSeriesName    string
		seriesPosition   string
		addSeries        bool
		addTags          bool
		pageCountStr     string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Add Book").
				Description("Add a new book to your collection.").
				Next(true).
				NextLabel("Next"),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Placeholder("required").
				Validate(huh.ValidateNotEmpty()).
				Value(&input.Title),
			huh.NewInput().
				Title("Headline").
				Placeholder("optional").
				Value(&input.Headline),
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
				Placeholder("required — e.g. 1963 or 1963-11-22").
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
			huh.NewSelect[int]().
				Title("Author").
				Options(authorOptions...).
				Value(&selectedAuthorID),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Author Name").
				Placeholder("required").
				Validate(huh.ValidateNotEmpty()).
				Value(&authorName),
			huh.NewInput().
				Title("Author Sort Name").
				PlaceholderFunc(func() string {
					return sortName(authorName)
				}, &authorName).
				Value(&authorSortName),
		).WithHideFunc(func() bool {
			return selectedAuthorID != 0
		}),

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
		).WithHideFunc(func() bool {
			return !addSeries
		}),

		huh.NewGroup(
			huh.NewInput().
				Title("Series Name").
				Placeholder("required").
				Validate(huh.ValidateNotEmpty()).
				Value(&newSeriesName),
		).WithHideFunc(func() bool {
			return !addSeries || selectedSeriesID != 0
		}),

		huh.NewGroup(
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
				Options(tagOptions...).
				Value(&input.TagIDs),
		).WithHideFunc(func() bool {
			return !addTags
		}),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Shelf Status").
				Options(
					huh.NewOption("On my shelf", "shelf"),
					huh.NewOption("Wishlist", "wishlist"),
					huh.NewOption("None", ""),
				).
				Value(&input.ShelfStatus),
		),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Format").
				Options(huh.NewOptions("hardback", "paperback", "ebook", "audiobook")...).
				Value(&input.Format),
			huh.NewInput().
				Title("Page Count").
				Placeholder("optional").
				Value(&pageCountStr),
			huh.NewInput().
				Title("Language").
				PlaceholderFunc(func() string {
					if input.OriginalLanguage != "" {
						return input.OriginalLanguage
					}
					return "english"
				}, &input.OriginalLanguage).
				Value(&input.Language),
			huh.NewInput().
				Title("ISBN").
				Placeholder("optional").
				Value(&input.ISBN),
		).WithHideFunc(func() bool {
			return input.ShelfStatus != "shelf"
		}),

		huh.NewGroup(
			huh.NewInput().
				Title("Cover Image").
				PlaceholderFunc(func() string {
					name := coverImageName(input.Title)
					if name == "" {
						return "optional"
					}
					return name
				}, &input.Title).
				Value(&input.CoverImage),
			huh.NewSelect[string]().
				Title("Source").
				Options(
					huh.NewOption("New", "new"),
					huh.NewOption("Second-hand", "second-hand"),
					huh.NewOption("Borrowed", "borrowed"),
					huh.NewOption("Gifted", "gifted"),
					huh.NewOption("Library", "library"),
				).
				Value(&input.Source),
			huh.NewInput().
				Title("Date Acquired").
				Placeholder("today").
				Value(&input.DateAcquired),
		).WithHideFunc(func() bool {
			return input.ShelfStatus != "shelf"
		}),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	if input.OriginalLanguage == "" {
		input.OriginalLanguage = "english"
	}

	if selectedAuthorID != 0 {
		for _, a := range authors {
			if a.ID == selectedAuthorID {
				input.Authors = append(input.Authors, repo.AuthorInput{
					ID:   a.ID,
					Name: a.Name,
					Role: "author",
				})
				break
			}
		}
	} else {
		if authorSortName == "" {
			authorSortName = sortName(authorName)
		}
		input.Authors = append(input.Authors, repo.AuthorInput{
			Name:     authorName,
			SortName: authorSortName,
			Role:     "author",
		})
	}

	if addSeries {
		pos, _ := strconv.ParseFloat(seriesPosition, 64)
		if selectedSeriesID != 0 {
			input.Series = &repo.SeriesInput{
				ID:       selectedSeriesID,
				Position: pos,
			}
		} else {
			input.Series = &repo.SeriesInput{
				Name:     newSeriesName,
				Position: pos,
			}
		}
	}

	if pageCountStr != "" {
		if n, err := strconv.Atoi(pageCountStr); err == nil {
			input.PageCount = n
		}
	}

	if input.CoverImage == "" {
		input.CoverImage = coverImageName(input.Title)
	}

	if input.Language == "" {
		if input.OriginalLanguage != "" {
			input.Language = input.OriginalLanguage
		} else {
			input.Language = "english"
		}
	}

	return nil
}

func sortName(authorName string) string {
	parts := strings.Fields(authorName)
	if len(parts) >= 2 {
		last := parts[len(parts)-1]
		rest := strings.Join(parts[:len(parts)-1], " ")
		return last + ", " + rest
	} else {
		return authorName
	}
}

func coverImageName(title string) string {
	if title == "" {
		return ""
	}
	name := strings.ToLower(title)
	name = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return ' '
	}, name)
	parts := strings.Fields(name)
	return strings.Join(parts, "_") + ".jpg"
}
