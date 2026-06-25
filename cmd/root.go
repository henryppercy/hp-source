package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/henryppercy/hp-source/internal/database"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

type app struct {
	repo *repo.Repo
	db   *sql.DB
}

func (a *app) open(path string) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
	}

	db, err := database.NewDB(path)
	if err != nil {
		return err
	}
	a.db = db
	a.repo = repo.New(db)
	return nil
}

func (a *app) close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

func defaultDBPath() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, "hp", "data.sqlite"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "hp", "data.sqlite"), nil
}

func resolveDBPath(flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	if env := os.Getenv("HP_DB"); env != "" {
		return env, nil
	}
	return defaultDBPath()
}

func newRootCmd() *cobra.Command {
	a := &app{}
	var dbFlag string

	root := &cobra.Command{
		Use:          "hp",
		Short:        "Personal data manager",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveDBPath(dbFlag)
			if err != nil {
				return err
			}
			return a.open(path)
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return a.close()
		},
	}

	root.PersistentFlags().StringVar(&dbFlag, "db", "", "database path (default: XDG data dir)")

	root.AddCommand(
		newMigrateCmd(a),
		newBookCmd(a),
		newReadCmd(a),
		newExportCmd(a),
		newSiteCmd(a),
		newPostCmd(a),
	)

	return root
}

func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
