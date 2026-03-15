package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/henryppercy/hp-source/internal/database"
	"github.com/henryppercy/hp-source/internal/forms"
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

			if err := database.Fresh(db); err != nil {
				return err
			}
		}

		return database.Migrate(db)
	},
}

var bookCmd = &cobra.Command{
	Use:   "book",
	Short: "Manage your book collection",
}

var bookAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new book",
	RunE: func(cmd *cobra.Command, args []string) error {
		mockGenres := []forms.Genre{
			{ID: 1, Name: "literary"},
			{ID: 2, Name: "thriller"},
			{ID: 3, Name: "mystery"},
			{ID: 4, Name: "science fiction"},
			{ID: 5, Name: "fantasy"},
			{ID: 6, Name: "horror"},
			{ID: 7, Name: "romance"},
			{ID: 8, Name: "historical"},
			{ID: 9, Name: "adventure"},
			{ID: 10, Name: "comedy"},
			{ID: 11, Name: "short stories"},
			{ID: 12, Name: "biography"},
			{ID: 13, Name: "history"},
			{ID: 14, Name: "science"},
			{ID: 15, Name: "philosophy"},
			{ID: 16, Name: "politics"},
			{ID: 17, Name: "self-help"},
		}

		mockAuthors := []forms.Author{
			{ID: 1, Name: "Frank Herbert", SortName: "Herbert, Frank"},
			{ID: 2, Name: "Ursula K. Le Guin", SortName: "Le Guin, Ursula K."},
			{ID: 3, Name: "George Orwell", SortName: "Orwell, George"},
		}

		mockTags := []forms.Tag{
			{ID: 1, Name: "dystopian"},
			{ID: 2, Name: "classic"},
			{ID: 3, Name: "space opera"},
			{ID: 4, Name: "cyberpunk"},
			{ID: 5, Name: "coming of age"},
		}

		mockSeries := []forms.Series{
			{ID: 1, Name: "Dune"},
			{ID: 2, Name: "Earthsea"},
			{ID: 3, Name: "The Lord of the Rings"},
		}

		input := &forms.BookInput{}
		if err := forms.AddBook(
			input,
			mockGenres,
			mockAuthors,
			mockTags,
			mockSeries,
		); err != nil {
			return err
		}

		// later: pass input to repo
		fmt.Printf("%+v\n", input)
		return nil
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

	bookCmd.AddCommand(bookAddCmd)
	rootCmd.AddCommand(bookCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
