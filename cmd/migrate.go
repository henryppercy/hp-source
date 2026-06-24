package cmd

import (
	"fmt"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/henryppercy/hp-source/internal/database"
	"github.com/spf13/cobra"
)

func newMigrateCmd(a *app) *cobra.Command {
	var fresh bool

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMigrate(a, fresh)
		},
	}

	cmd.Flags().BoolVar(&fresh, "fresh", false, "Drop all tables and re-run migrations")
	return cmd
}

func runMigrate(a *app, fresh bool) error {
	if fresh {
		var confirm bool

		err := huh.NewConfirm().
			WithButtonAlignment(lipgloss.Left).
			Title("Drop all tables and re-run migrations?").
			Affirmative("Yes").
			Negative("Cancel").
			Value(&confirm).
			Run()
		if err != nil {
			return err
		}

		if !confirm {
			fmt.Println("aborted")
			return nil
		}

		if err := database.Fresh(a.db); err != nil {
			return err
		}
	}

	applied, err := database.Migrate(a.db)
	if err != nil {
		return err
	}

	for _, name := range applied {
		fmt.Printf("applied: %s\n", name)
	}
	if len(applied) == 0 {
		fmt.Println("database up to date")
	}
	return nil
}
