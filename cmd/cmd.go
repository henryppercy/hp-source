package cmd

import (
	"fmt"
	"os"

	"github.com/henryppercy/hp-source/internal/database"
)

func Run(args []string) {
	db, err := database.NewDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cmd %w:", err)
		os.Exit(1)
	}
	defer db.Close()

	// pass to repo
}
