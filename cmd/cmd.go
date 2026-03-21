package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export reads as mdx files",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir, _ := cmd.Flags().GetString("out")

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		reads, err := r.ListExportReads()
		if err != nil {
			return err
		}

		for _, read := range reads {
			filename := filepath.Join(outputDir, read.Slug()+".mdx")
			content := buildMdx(read)

			if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", filename, err)
			}

			fmt.Printf("exported: %s\n", filename)
		}

		fmt.Printf("\n%d reads exported.\n", len(reads))
		return nil
	},
}

func buildMdx(e repo.ExportRead) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "---\n")
	fmt.Fprintf(&sb, "author: \"%s\"\n", e.Author)
	fmt.Fprintf(&sb, "title: \"%s\"\n", e.Title)
	fmt.Fprintf(&sb, "headline: \"%s\"\n", e.Headline)
	fmt.Fprintf(&sb, "series: \"%s\"\n", e.SeriesName)

	if e.SeriesPosition > 0 {
		fmt.Fprintf(&sb, "series_pos: %g\n", e.SeriesPosition)
	} else {
		fmt.Fprintf(&sb, "series_pos:\n")
	}

	fmt.Fprintf(&sb, "image_url: \"/%s\"\n", e.CoverImage)
	fmt.Fprintf(&sb, "date_published: %s\n", e.DatePublished)
	fmt.Fprintf(&sb, "date_started: %s\n", e.DateStarted)
	fmt.Fprintf(&sb, "date_finished: %s\n", e.DateFinished)

	if e.Rating > 0 {
		fmt.Fprintf(&sb, "rating: %s\n", e.RatingDisplay())
	} else {
		fmt.Fprintf(&sb, "rating:\n")
	}

	fmt.Fprintf(&sb, "type: \"%s\"\n", e.BookType)
	fmt.Fprintf(&sb, "genre: \"%s\"\n", e.Genre)

	if len(e.Tags) > 0 {
		fmt.Fprintf(&sb, "tags:\n")
		for _, tag := range e.Tags {
			fmt.Fprintf(&sb, "  - \"%s\"\n", tag)
		}
	} else {
		fmt.Fprintf(&sb, "tags:\n")
	}

	fmt.Fprintf(&sb, "format: \"%s\"\n", e.Format)
	fmt.Fprintf(&sb, "language: \"%s\"\n", e.Language)
	fmt.Fprintf(&sb, "original_language: \"%s\"\n", e.OriginalLanguage)

	if e.PageCount > 0 {
		fmt.Fprintf(&sb, "page_count: %d\n", e.PageCount)
	} else {
		fmt.Fprintf(&sb, "page_count:\n")
	}

	fmt.Fprintf(&sb, "---\n")

	return sb.String()
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

	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().String("out", "./export", "Output directory for mdx files")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
