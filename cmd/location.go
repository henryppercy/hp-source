package cmd

import (
	"fmt"

	"github.com/henryppercy/hp-source/internal/forms"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

func newLocationCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "location",
		Short: "Manage places",
	}
	cmd.AddCommand(
		newLocationAddCmd(a),
		newLocationListCmd(a),
	)
	return cmd
}

func newLocationAddCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a place",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLocationAdd(a.repo)
		},
	}
}

func newLocationListCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List places",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLocationList(a.repo)
		},
	}
}

func runLocationAdd(r *repo.Repo) error {
	input := &repo.LocationInput{}
	if err := forms.AddLocation(input); err != nil {
		return err
	}

	if err := r.CreateLocation(input); err != nil {
		return err
	}

	fmt.Printf("added '%s'.\n", input.Name)
	return nil
}

func runLocationList(r *repo.Repo) error {
	locations, err := r.ListLocations()
	if err != nil {
		return err
	}
	if len(locations) == 0 {
		fmt.Println("No locations yet.")
		return nil
	}
	for _, l := range locations {
		code := l.Code
		if code == "" {
			code = "--"
		}
		fmt.Printf("%-20s %-4s %s\n", l.Name, code, l.Slug)
	}
	return nil
}
