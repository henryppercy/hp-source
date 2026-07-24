package forms

import (
	"fmt"
	"strings"
	"time"

	huh "charm.land/huh/v2"
	"github.com/henryppercy/hp-source/internal/repo"
)

func ratingHuhOptions() []huh.Option[int] {
	ratings := repo.RatingOptions()
	opts := make([]huh.Option[int], len(ratings))
	for i, r := range ratings {
		opts[i] = huh.NewOption(r.Label, r.Value)
	}
	return opts
}

func LogRead(input *repo.ReadInput, books []repo.BookSummary, fetchCopies func(int) ([]repo.Copy, error)) error {
	bookOptions := make([]huh.Option[int], len(books))
	for i, b := range books {
		label := b.Title
		if b.Author != "" {
			label += " - " + b.Author
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
				Options(ratingHuhOptions()...).
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
								label += " - " + b.Author
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
						fmt.Fprintf(&sb, "Rating:     %s/5\n", repo.RatingDisplay(input.Rating))
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

	return nil
}

func StartRead(input *repo.StartReadInput, books []repo.BookSummary, fetchCopies func(int) ([]repo.Copy, error)) error {
	bookOptions := make([]huh.Option[int], len(books))
	for i, b := range books {
		label := b.Title
		if b.Author != "" {
			label += " - " + b.Author
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
				Title("Start Read").
				Description("Begin reading a book.").
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
			huh.NewInput().
				Title("Date Started").
				Placeholder(today).
				Value(&input.DateStarted),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Confirm").
				DescriptionFunc(func() string {
					var sb strings.Builder

					for _, b := range books {
						if b.ID == input.BookID {
							label := b.Title
							if b.Author != "" {
								label += " - " + b.Author
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

					started := input.DateStarted
					if started == "" {
						started = "today"
					}
					fmt.Fprintf(&sb, "Started:    %s\n", started)

					return sb.String()
				}, &input.DateStarted).
				Next(true).
				NextLabel("Save"),
		),
	)

	err := form.Run()
	if err != nil {
		return err
	}

	if input.DateStarted == "" {
		input.DateStarted = today
	}

	return nil
}

func EditRead(input *repo.EditReadInput, reads []repo.ReadDetail) error {
	readOptions := make([]huh.Option[int], len(reads))
	for i, r := range reads {
		label := r.BookTitle
		if r.Author != "" {
			label += " - " + r.Author
		}
		label += " (" + r.Status + ")"
		readOptions[i] = huh.NewOption(label, r.ReadID)
	}

	readSelect := huh.NewSelect[int]().
		Title("Edit Read").
		Description("Which read do you want to edit?").
		Options(readOptions...).
		Value(&input.ReadID)

	if len(reads) > 10 {
		readSelect.Height(10)
	}

	picker := huh.NewForm(huh.NewGroup(readSelect))
	if err := picker.Run(); err != nil {
		return err
	}

	var chosen repo.ReadDetail
	for _, r := range reads {
		if r.ReadID == input.ReadID {
			chosen = r
			input.Status = r.Status
			input.Rating = r.Rating
			input.DateStarted = r.DateStarted
			input.DateFinished = r.DateFinished
			break
		}
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Status").
				Options(
					huh.NewOption("Reading", "reading"),
					huh.NewOption("Finished", "finished"),
					huh.NewOption("Abandoned", "abandoned"),
				).
				Value(&input.Status),
		),

		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Rating").
				Options(ratingHuhOptions()...).
				Value(&input.Rating),
		).WithHideFunc(func() bool {
			return input.Status == "abandoned"
		}),

		huh.NewGroup(
			huh.NewInput().
				Title("Date Started").
				Validate(validateDateOptional).
				Value(&input.DateStarted),
			huh.NewInput().
				Title("Date Finished").
				Validate(validateDateOptional).
				Value(&input.DateFinished),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Confirm Changes").
				DescriptionFunc(func() string {
					var sb strings.Builder

					label := chosen.BookTitle
					if chosen.Author != "" {
						label += " - " + chosen.Author
					}
					fmt.Fprintf(&sb, "Book:       %s\n", label)

					if chosen.Format != "" {
						fmt.Fprintf(&sb, "Copy:       %s\n", chosen.Format)
					}

					fmt.Fprintf(&sb, "Status:     %s\n", input.Status)

					if input.Status != "abandoned" && input.Rating > 0 {
						fmt.Fprintf(&sb, "Rating:     %s/5\n", repo.RatingDisplay(input.Rating))
					}

					if input.DateStarted != "" {
						fmt.Fprintf(&sb, "Started:    %s\n", input.DateStarted)
					}
					if input.DateFinished != "" {
						fmt.Fprintf(&sb, "Finished:   %s\n", input.DateFinished)
					}

					return sb.String()
				}, &input.DateFinished).
				Next(true).
				NextLabel("Save"),
		),
	)

	return form.Run()
}

func FinishRead(input *repo.FinishReadInput, reads []repo.ActiveRead) error {
	readOptions := make([]huh.Option[int], len(reads))
	for i, r := range reads {
		label := r.BookTitle
		if r.Author != "" {
			label += " - " + r.Author
		}
		if r.Format != "" {
			label += " (" + r.Format + ")"
		}
		readOptions[i] = huh.NewOption(label, r.ReadID)
	}

	today := time.Now().Format("2006-01-02")

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Finish Read").
				Description("Mark a read as finished or abandoned.").
				Next(true).
				NextLabel("Next"),
		),

		huh.NewGroup(
			huh.NewSelect[int]().
				Title("Which book?").
				Options(readOptions...).
				Value(&input.ReadID),
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
				Options(ratingHuhOptions()...).
				Value(&input.Rating),
		).WithHideFunc(func() bool {
			return input.Status == "abandoned"
		}),

		huh.NewGroup(
			huh.NewInput().
				Title("Date Finished").
				Placeholder(today).
				Value(&input.DateFinished),
		),

		huh.NewGroup(
			huh.NewNote().
				Title("Confirm").
				DescriptionFunc(func() string {
					var sb strings.Builder

					for _, r := range reads {
						if r.ReadID == input.ReadID {
							label := r.BookTitle
							if r.Author != "" {
								label += " - " + r.Author
							}
							fmt.Fprintf(&sb, "Book:       %s\n", label)
							break
						}
					}

					fmt.Fprintf(&sb, "Status:     %s\n", input.Status)

					if input.Status != "abandoned" && input.Rating > 0 {
						fmt.Fprintf(&sb, "Rating:     %s/5\n", repo.RatingDisplay(input.Rating))
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

	return nil
}
