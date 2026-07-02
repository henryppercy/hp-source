package cmd

import (
	"github.com/henryppercy/hp-source/internal/site"
	"github.com/spf13/cobra"
)

func newSiteCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "site",
		Short: "Build and serve the static site",
	}
	cmd.AddCommand(
		newSiteBuildCmd(a),
		newSiteServeCmd(a),
	)
	return cmd
}

func newSiteBuildCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the static site",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, _ := cmd.Flags().GetString("out")
			from, _ := cmd.Flags().GetString("from")

			id, err := site.RecordBuild(a.repo, from)
			if err != nil {
				return err
			}
			if err := site.Build(a.repo, out); err != nil {
				return err
			}
			return a.repo.MarkBuildSuccess(id)
		},
	}
	cmd.Flags().String("out", "./dist", "Output directory for the built site")
	cmd.Flags().String("from", site.HomeSlug, "Location slug the build is filed from")
	return cmd
}

func newSiteServeCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Build and serve the static site locally",
		RunE: func(cmd *cobra.Command, args []string) error {
			out, _ := cmd.Flags().GetString("out")
			addr, _ := cmd.Flags().GetString("addr")
			watch, _ := cmd.Flags().GetBool("watch")
			return site.Serve(a.repo, out, addr, watch)
		},
	}
	cmd.Flags().String("out", "./dist", "Output directory for the built site")
	cmd.Flags().String("addr", ":8080", "Address to serve on")
	cmd.Flags().Bool("watch", false, "Rebuild on template/asset changes (reads from internal/site on disk)")
	return cmd
}
