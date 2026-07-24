package cmd

import (
	"fmt"

	"github.com/henryppercy/hp-source/internal/forms"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

func newAuthorCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "author",
		Short: "Manage authors",
	}
	cmd.AddCommand(newAuthorAddCmd(a))
	return cmd
}

func newAuthorAddCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a new author",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthorAdd(a.repo)
		},
	}
}

func runAuthorAdd(r *repo.Repo) error {
	var name, sort string
	if err := forms.AddAuthor(&name, &sort); err != nil {
		return err
	}
	if _, err := r.CreateAuthor(name, sort); err != nil {
		return err
	}
	fmt.Println("Author added.")
	return nil
}
