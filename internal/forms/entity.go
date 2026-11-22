package forms

import huh "charm.land/huh/v2"

// AddAuthor collects an author's details into name and sort.
func AddAuthor(name, sort *string) error {
	return personForm("Add Author", name, sort)
}

// AddTranslator collects a translator's details into name and sort.
func AddTranslator(name, sort *string) error {
	return personForm("Add Translator", name, sort)
}

func personForm(title string, name, sort *string) error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().Title(title).Next(true).NextLabel("Next"),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Name").
				Placeholder("required").
				Validate(huh.ValidateNotEmpty()).
				Value(name),
			huh.NewInput().
				Title("Sort Name").
				PlaceholderFunc(func() string {
					return sortName(*name)
				}, name).
				Value(sort),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}
	if *sort == "" {
		*sort = sortName(*name)
	}
	return nil
}

// AddSeries collects a series name into name.
func AddSeries(name *string) error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Series Name").
				Placeholder("required").
				Validate(huh.ValidateNotEmpty()).
				Value(name),
		),
	)
	return form.Run()
}
