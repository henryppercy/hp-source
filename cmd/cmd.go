package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/henryppercy/hp-source/internal/database"
	"github.com/henryppercy/hp-source/internal/forms"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

var db *sql.DB
var r *repo.Repo
var freshMigrate bool

var rootCmd = &cobra.Command{
	Use:   "hp",
	Short: "Personal data manager",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		db, err = database.NewDB()
		if err != nil {
			return err
		}
		r = repo.New(db)
		return nil
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
		if err := forms.AddBook(
			input,
			genres,
			authors,
			tags,
			series,
		); err != nil {
			return err
		}

		err = r.AddBook(input)
		if err != nil {
			return err
		}

		return nil
	},
}

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Manage your reading",
}

var readLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Log a completed read",
	RunE: func(cmd *cobra.Command, args []string) error {
		books, err := r.ListBooks(true)
		if err != nil {
			return err
		}

		input := &repo.ReadInput{}
		if err := forms.LogRead(input, books, r.ListCopies); err != nil {
			return err
		}

		if err := r.AddRead(input); err != nil {
			return err
		}

		return nil
	},
}

var readStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start reading a book",
	RunE: func(cmd *cobra.Command, args []string) error {
		books, err := r.ListBooksAvailableToRead()
		if err != nil {
			return err
		}

		if len(books) == 0 {
			fmt.Println("No books available to start reading.")
			return nil
		}

		input := &repo.StartReadInput{}
		if err := forms.StartRead(input, books, r.ListCopies); err != nil {
			return err
		}

		if err := r.StartRead(input.BookID, input.CopyID, input.DateStarted); err != nil {
			return err
		}

		fmt.Println("Read started.")
		return nil
	},
}

var readFinishCmd = &cobra.Command{
	Use:   "finish",
	Short: "Finish or abandon a read",
	RunE: func(cmd *cobra.Command, args []string) error {
		reads, err := r.ListActiveReads()
		if err != nil {
			return err
		}

		if len(reads) == 0 {
			fmt.Println("No active reads to finish.")
			return nil
		}

		input := &repo.FinishReadInput{}
		if err := forms.FinishRead(input, reads); err != nil {
			return err
		}

		if err := r.FinishRead(input.ReadID, input.Status, input.Rating, input.DateFinished); err != nil {
			return err
		}

		fmt.Println("Read updated.")
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

	readCmd.AddCommand(readStartCmd)
	readCmd.AddCommand(readFinishCmd)
	readCmd.AddCommand(readLogCmd)
	rootCmd.AddCommand(readCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
