package forms

import (
	"fmt"
	"strconv"
	"strings"

	huh "charm.land/huh/v2"
	"github.com/henryppercy/hp-source/internal/repo"
)

// LogProgress records a page, a note, or both against an active read.
func LogProgress(input *repo.ReadLogInput, reads []repo.ActiveRead) error {
	readOptions := make([]huh.Option[int], len(reads))
	for i, rd := range reads {
		label := rd.BookTitle
		if rd.Author != "" {
			label += " - " + rd.Author
		}
		readOptions[i] = huh.NewOption(label, rd.ReadID)
	}

	readSelect := huh.NewSelect[int]().
		Title("Read").
		Options(readOptions...).
		Value(&input.ReadID)

	if len(reads) > 10 {
		readSelect.Height(10)
	}

	page := ""

	form := huh.NewForm(
		huh.NewGroup(readSelect),
		huh.NewGroup(
			huh.NewInput().
				Title("Page").
				Placeholder("optional").
				Validate(validateOptionalPage).
				Value(&page),
			huh.NewText().
				Title("Note").
				Placeholder("optional").
				Value(&input.Note),
		),
		huh.NewGroup(
			huh.NewNote().
				Title("Confirm Entry").
				DescriptionFunc(func() string {
					var sb strings.Builder
					for _, rd := range reads {
						if rd.ReadID == input.ReadID {
							fmt.Fprintf(&sb, "Read:       %s\n", rd.BookTitle)
							break
						}
					}
					if page != "" {
						fmt.Fprintf(&sb, "Page:       %s\n", page)
					}
					if input.Note != "" {
						fmt.Fprintf(&sb, "\n%s\n", input.Note)
					}
					return sb.String()
				}, &input.Note).
				Next(true).
				NextLabel("Save"),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if page != "" {
		input.Page, _ = strconv.Atoi(page)
	}

	if input.Page == 0 && input.Note == "" {
		return fmt.Errorf("nothing to log: give a page, a note, or both")
	}

	return nil
}

func validateOptionalPage(s string) error {
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return fmt.Errorf("must be a positive page number")
	}
	return nil
}
