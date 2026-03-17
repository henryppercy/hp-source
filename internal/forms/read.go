package forms

import (
	"fmt"
	"strings"
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

		huh.NewGroup(
			huh.NewNote().
				Title("Confirm Read").
				DescriptionFunc(func() string {
					var sb strings.Builder

					for _, b := range books {
						if b.ID == input.BookID {
							label := b.Title
							if b.Author != "" {
								label += " — " + b.Author
							}
							fmt.Fprintf(&sb, "Book:       %s\n", label)
							break
						}
					}

					copies, err := fetchCopies(input.BookID)
					if err == nil {
						for _, c := range copies {
							if c.ID == input.CopyID {
								fmt.Fprintf(&sb, "Copy:       %s\n", c.Format)
								break
							}
						}
					}

					fmt.Fprintf(&sb, "Status:     %s\n", input.Status)

					if input.Status != "abandoned" && input.Rating > 0 {
						ratings := map[int]string{
							10: "5", 9: "4.5", 8: "4", 7: "3.5", 6: "3",
							5: "2.5", 4: "2", 3: "1.5", 2: "1", 1: "0.5",
						}
						fmt.Fprintf(&sb, "Rating:     %s/5\n", ratings[input.Rating])
					}

					if input.DateStarted != "" {
						fmt.Fprintf(&sb, "Started:    %s\n", input.DateStarted)
					}

					finished := input.DateFinished
					if finished == "" {
						finished = "today"
					}
					fmt.Fprintf(&sb, "Finished:   %s\n", finished)

					return sb.String()
				}, &input.DateFinished).
				Next(true).
				NextLabel("Save"),
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
