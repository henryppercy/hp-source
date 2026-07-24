package cmd

import (
	"fmt"

	"github.com/henryppercy/hp-source/internal/forms"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

func newSeriesCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "series",
		Short: "Manage series",
	}
	cmd.AddCommand(newSeriesAddCmd(a))
	return cmd
}

func newSeriesAddCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a new series",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSeriesAdd(a.repo)
		},
	}
}

func runSeriesAdd(r *repo.Repo) error {
	var name string
	if err := forms.AddSeries(&name); err != nil {
		return err
	}
	if _, err := r.CreateSeries(name); err != nil {
		return err
	}
	fmt.Println("Series added.")
	return nil
}
