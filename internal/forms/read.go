package forms

import (
	"time"

	huh "charm.land/huh/v2"
	"github.com/henryppercy/hp-source/internal/repo"
)

func LogRead(input *repo.ReadInput, books []repo.BookSummary, fetchCopies func(int) ([]repo.Copy, error)) error {
	bookOptions := make([]huh.Option[int], len(books))
	for i, b := range books {
		label := b.Title
		if b.Author != "" {
			label += " — " + b.Author
		}
		bookOptions[i] = huh.NewOption(label, b.ID)
	}

	today := time.Now().Format("2006-01-02")

	bookSelect := huh.NewSelect[int]().
		Title("Book").
		Options(bookOptions...).
		Value(&input.BookID)

	if len(books) > 10 {
		bookSelect.Height(10)
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Log Read").
				Description("Record a completed read.").
				Next(true).
				NextLabel("Next"),
		),

		huh.NewGroup(
			bookSelect,
		),

		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Copy").
				OptionsFunc(func() []huh.Option[int] {
					copies, err := fetchCopies(input.BookID)
					if err != nil || len(copies) == 0 {
						return []huh.Option[int]{huh.NewOption("No copies found", 0)}
					}
					opts := make([]huh.Option[int], len(copies))
					for i, c := range copies {
						opts[i] = huh.NewOption(c.Format, c.ID)
					}
					return opts
				}, &input.BookID).
				Value(&input.CopyID),
			huh.NewSelect[string]().
				Title("Status").
				Options(
					huh.NewOption("Finished", "finished"),
					huh.NewOption("Abandoned", "abandoned"),
				).
				Value(&input.Status),
		),
		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Rating").
				Options(
					huh.NewOption("5", 10),
					huh.NewOption("4.5", 9),
					huh.NewOption("4", 8),
					huh.NewOption("3.5", 7),
					huh.NewOption("3", 6),
					huh.NewOption("2.5", 5),
					huh.NewOption("2", 4),
					huh.NewOption("1.5", 3),
					huh.NewOption("1", 2),
					huh.NewOption("0.5", 1),
				).
				Value(&input.Rating),
		).WithHideFunc(func() bool {
			return input.Status == "abandoned"
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Date Started").
				Validate(validateDate).
				Value(&input.DateStarted),
			huh.NewInput().
				Title("Date Finished").
				Placeholder(today).
				Value(&input.DateFinished),
		),
	)
	err := form.Run()
	if err != nil {
		return err
	}

	if input.DateFinished == "" {
		input.DateFinished = today
	}

	if input.Rating == 0 {
		input.Rating = 0
	}

	return nil
}
