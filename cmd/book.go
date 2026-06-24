package cmd

import (
	"github.com/henryppercy/hp-source/internal/forms"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

func newBookCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "book",
		Short: "Manage your book collection",
	}
	cmd.AddCommand(newBookAddCmd(a))
	return cmd
}

func newBookAddCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a new book",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBookAdd(a.repo)
		},
	}
}

func runBookAdd(r *repo.Repo) error {
	genres, err := r.ListGenres()
	if err != nil {
		return err
	}
	authors, err := r.ListAuthors()
	if err != nil {
		return err
	}
	tags, err := r.ListTags()
	if err != nil {
		return err
	}
	series, err := r.ListSeries()
	if err != nil {
		return err
	}

	input := &repo.BookInput{}
	if err := forms.AddBook(input, genres, authors, tags, series); err != nil {
		return err
	}

	return r.AddBook(input)
}
