package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/henryppercy/hp-source/internal/database"
	"github.com/spf13/cobra"
)

var db *sql.DB
var freshMigrate bool

var rootCmd = &cobra.Command{
	Use:   "hp",
	Short: "Personal data manager",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		db, err = database.NewDB()
		return err
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if db != nil {
			db.Close()
		}
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if freshMigrate {
			var confirm bool

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						WithButtonAlignment(lipgloss.Left).
						Title("Drop all tables and re-run migrations?").
						Affirmative("Yes").
						Negative("Cancel").
						Value(&confirm),
				),
			)

			err := form.Run()
			if err != nil {
				return err
			}

			if !confirm {
				fmt.Println("aborted")
				return nil
			}

			if err := database.Fresh(db); err != nil {
				return err
			}
		}
		return database.Migrate(db)
	},
}

func Execute() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().BoolVar(
		&freshMigrate,
		"fresh",
		false,
		"Drop all tables and re-run migrations",
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
