package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

func newExportCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export reads as mdx files",
		RunE: func(cmd *cobra.Command, args []string) error {
			outputDir, _ := cmd.Flags().GetString("out")
			return runExport(a.repo, outputDir)
		},
	}

	cmd.Flags().String("out", "./export", "Output directory for mdx files")
	return cmd
}

func runExport(r *repo.Repo, outputDir string) error {
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
