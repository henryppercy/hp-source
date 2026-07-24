package forms

import (
	"fmt"
	"strconv"
	"strings"

	huh "charm.land/huh/v2"
	"github.com/henryppercy/hp-source/internal/repo"
)

// SelectBook picks a book into the given pointer.
func SelectBook(books []repo.BookSummary, bookID *int) error {
	options := make([]huh.Option[int], len(books))
	for i, b := range books {
		label := b.Title
		if b.Author != "" {
			label += " - " + b.Author
		}
		options[i] = huh.NewOption(label, b.ID)
	}

	sel := huh.NewSelect[int]().Title("Book").Options(options...).Value(bookID)
	if len(books) > 10 {
		sel.Height(10)
	}
	return huh.NewForm(huh.NewGroup(sel)).Run()
}

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
						opts[i] = huh.NewOption(copyLabel(c), c.ID)
					}
					return opts
				}, bookID).
				Value(copyID),
		),
	)

	return form.Run()
}

// copyLabel distinguishes copies of one book by format, language and edition title.
func copyLabel(c repo.Copy) string {
	label := c.Format
	if c.Language != "" {
		label += " - " + c.Language
	}
	if c.Title != "" {
		label += " - " + c.Title
	}
	return label
}

// CopyForm edits a copy's details, shared by book add, copy add and copy edit.
// defaultTitle prefills a blank title (the work's title); originalLanguage seeds
// the language placeholder.
func CopyForm(in *repo.CopyInput, translators []repo.Translator, defaultTitle, originalLanguage string) error {
	if in.Title == "" {
		in.Title = defaultTitle
	}

	pageCount := ""
	if in.PageCount > 0 {
		pageCount = strconv.Itoa(in.PageCount)
	}

	translatorOptions := make([]huh.Option[int], len(translators))
	for i, t := range translators {
		translatorOptions[i] = huh.NewOption(t.Name, t.ID)
	}

	var selectedTranslatorIDs []int
	for _, t := range in.Translators {
		if t.ID != 0 {
			selectedTranslatorIDs = append(selectedTranslatorIDs, t.ID)
		}
	}
	addTranslators := len(in.Translators) > 0

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Copy").
				Description("The specific edition, in its own language and title.").
				Next(true).
				NextLabel("Next"),
		),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Shelf Status").
				Options(
					huh.NewOption("On my shelf", "shelf"),
					huh.NewOption("Wishlist", "wishlist"),
					huh.NewOption("None", ""),
				).
				Value(&in.ShelfStatus),
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
				PlaceholderFunc(func() string {
					if originalLanguage != "" {
						return originalLanguage
					}
					return "english"
				}, &originalLanguage).
				Value(&in.Language),
			huh.NewInput().
				Title("ISBN").
				Placeholder("optional").
				Value(&in.ISBN),
		),

		huh.NewGroup(
			huh.NewInput().
				Title("Cover Image").
				PlaceholderFunc(func() string {
					name := coverImageName(in.Title)
					if name == "" {
						return "optional"
					}
					return name
				}, &in.Title).
				Value(&in.CoverImage),
			huh.NewSelect[string]().
				Title("Source").
				Options(
					huh.NewOption("None", ""),
					huh.NewOption("Bought", "bought"),
					huh.NewOption("Gifted", "gifted"),
					huh.NewOption("Borrowed", "borrowed"),
					huh.NewOption("Library", "library"),
				).
				Value(&in.Source),
			huh.NewConfirm().
				Title("Second-hand?").
				Affirmative("Yes").
				Negative("No").
				Value(&in.SecondHand),
			huh.NewInput().
				Title("Date Acquired").
				Placeholder("YYYY-MM-DD").
				Validate(validateDateOptional).
				Value(&in.DateAcquired),
		),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Add translators?").
				Affirmative("Yes").
				Negative("No").
				Value(&addTranslators),
		),

		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Translators").
				Height(15).
				Options(translatorOptions...).
				Value(&selectedTranslatorIDs),
		).WithHideFunc(func() bool {
			return !addTranslators
		}),

		huh.NewGroup(
			huh.NewNote().
				Title("Confirm Copy").
				DescriptionFunc(func() string {
					return copySummary(in, pageCount, originalLanguage)
				}, &in.Title).
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

	if in.Language == "" {
		if originalLanguage != "" {
			in.Language = originalLanguage
		} else {
			in.Language = "english"
		}
	}

	if in.CoverImage == "" {
		in.CoverImage = coverImageName(in.Title)
	}

	if !addTranslators {
		in.Translators = nil
		return nil
	}
	in.Translators = resolveTranslators(translators, selectedTranslatorIDs)
	return nil
}

// resolveTranslators builds the copy's translator list from the selected ids.
func resolveTranslators(translators []repo.Translator, selectedIDs []int) []repo.TranslatorInput {
	var out []repo.TranslatorInput
	for _, id := range selectedIDs {
		for _, t := range translators {
			if t.ID == id {
				out = append(out, repo.TranslatorInput{ID: t.ID, Name: t.Name, SortName: t.SortName})
				break
			}
		}
	}
	return out
}

func copySummary(in *repo.CopyInput, pageCount, originalLanguage string) string {
	var sb strings.Builder
	shelf := in.ShelfStatus
	if shelf == "" {
		shelf = "none"
	}
	fmt.Fprintf(&sb, "Shelf:      %s\n", shelf)
	fmt.Fprintf(&sb, "Title:      %s\n", in.Title)
	if in.Headline != "" {
		fmt.Fprintf(&sb, "Headline:   %s\n", in.Headline)
	}
	fmt.Fprintf(&sb, "Format:     %s\n", in.Format)
	if pageCount != "" {
		fmt.Fprintf(&sb, "Pages:      %s\n", pageCount)
	}
	lang := in.Language
	if lang == "" {
		if originalLanguage != "" {
			lang = originalLanguage
		} else {
			lang = "english"
		}
	}
	fmt.Fprintf(&sb, "Language:   %s\n", lang)
	if in.ISBN != "" {
		fmt.Fprintf(&sb, "ISBN:       %s\n", in.ISBN)
	}
	cover := in.CoverImage
	if cover == "" {
		cover = coverImageName(in.Title)
	}
	if cover != "" {
		fmt.Fprintf(&sb, "Cover:      %s\n", cover)
	}
	source := in.Source
	if source == "" {
		source = "none"
	}
	fmt.Fprintf(&sb, "Source:     %s\n", source)
	fmt.Fprintf(&sb, "Second-hand: %s\n", yesNo(in.SecondHand))
	acquired := in.DateAcquired
	if acquired == "" {
		acquired = "none"
	}
	fmt.Fprintf(&sb, "Acquired:   %s\n", acquired)
	return sb.String()
}
