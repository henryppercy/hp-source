package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/henryppercy/hp-source/internal/export"
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
		content := export.MDX(read)

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}

		fmt.Printf("exported: %s\n", filename)
	}

	fmt.Printf("\n%d reads exported.\n", len(reads))
	return nil
}
