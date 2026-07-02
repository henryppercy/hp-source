package forms

import (
	"fmt"
	"strconv"
	"strings"

	huh "charm.land/huh/v2"
	"github.com/henryppercy/hp-source/internal/repo"
)

// SelectCopy picks a book, then one of its copies, into the given pointers.
func SelectCopy(books []repo.BookSummary, fetchCopies func(int) ([]repo.Copy, error), bookID, copyID *int) error {
	bookOptions := make([]huh.Option[int], len(books))
	for i, b := range books {
		label := b.Title
		if b.Author != "" {
			label += " - " + b.Author
		}
		bookOptions[i] = huh.NewOption(label, b.ID)
	}

	bookSelect := huh.NewSelect[int]().
		Title("Book").
		Options(bookOptions...).
		Value(bookID)

	if len(books) > 10 {
		bookSelect.Height(10)
	}

	form := huh.NewForm(
		huh.NewGroup(bookSelect),
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Copy").
				OptionsFunc(func() []huh.Option[int] {
					copies, err := fetchCopies(*bookID)
					if err != nil || len(copies) == 0 {
						return []huh.Option[int]{huh.NewOption("No copies found", 0)}
					}
					opts := make([]huh.Option[int], len(copies))
					for i, c := range copies {
						opts[i] = huh.NewOption(c.Format, c.ID)
					}
					return opts
				}, bookID).
				Value(copyID),
		),
	)

	return form.Run()
}

// EditCopy edits a copy's details, prefilled from its current values.
func EditCopy(in *repo.CopyInput) error {
	pageCount := ""
	if in.PageCount > 0 {
		pageCount = strconv.Itoa(in.PageCount)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Edit Copy").
				Next(true).
				NextLabel("Next"),
		),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Format").
				Options(huh.NewOptions("hardback", "paperback", "ebook", "audiobook")...).
				Value(&in.Format),
			huh.NewInput().
				Title("Page Count").
				Placeholder("optional").
				Value(&pageCount),
			huh.NewInput().
				Title("Language").
				Validate(huh.ValidateNotEmpty()).
				Value(&in.Language),
			huh.NewInput().
				Title("ISBN").
				Placeholder("optional").
				Value(&in.ISBN),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Cover Image").
				Placeholder("optional").
				Value(&in.CoverImage),
			huh.NewSelect[string]().
				Title("Source").
				Options(
					huh.NewOption("None", ""),
					huh.NewOption("New", "new"),
					huh.NewOption("Second-hand", "second-hand"),
					huh.NewOption("Borrowed", "borrowed"),
					huh.NewOption("Gifted", "gifted"),
					huh.NewOption("Library", "library"),
				).
				Value(&in.Source),
			huh.NewInput().
				Title("Date Acquired").
				Placeholder("e.g. 2026-06-20").
				Validate(validateDateOptional).
				Value(&in.DateAcquired),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Confirm Copy").
				DescriptionFunc(func() string {
					var sb strings.Builder
					fmt.Fprintf(&sb, "Format:     %s\n", in.Format)
					if pageCount != "" {
						fmt.Fprintf(&sb, "Pages:      %s\n", pageCount)
					}
					fmt.Fprintf(&sb, "Language:   %s\n", in.Language)
					if in.ISBN != "" {
						fmt.Fprintf(&sb, "ISBN:       %s\n", in.ISBN)
					}
					if in.CoverImage != "" {
						fmt.Fprintf(&sb, "Cover:      %s\n", in.CoverImage)
					}
					source := in.Source
					if source == "" {
						source = "none"
					}
					fmt.Fprintf(&sb, "Source:     %s\n", source)
					acquired := in.DateAcquired
					if acquired == "" {
						acquired = "none"
					}
					fmt.Fprintf(&sb, "Acquired:   %s\n", acquired)
					return sb.String()
				}, &in.DateAcquired).
				Next(true).
				NextLabel("Save"),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	in.PageCount = 0
	if pageCount != "" {
		if n, err := strconv.Atoi(pageCount); err == nil {
			in.PageCount = n
		}
	}

	return nil
}
