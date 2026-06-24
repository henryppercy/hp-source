package cmd

import (
	"fmt"

	"github.com/henryppercy/hp-source/internal/forms"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

func newReadCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Manage your reading",
	}
	cmd.AddCommand(
		newReadStartCmd(a),
		newReadFinishCmd(a),
		newReadLogCmd(a),
	)
	return cmd
}

func newReadStartCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start reading a book",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReadStart(a.repo)
		},
	}
}

func newReadFinishCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "finish",
		Short: "Finish or abandon a read",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReadFinish(a.repo)
		},
	}
}

func newReadLogCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "Log a completed read",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReadLog(a.repo)
		},
	}
}

func runReadStart(r *repo.Repo) error {
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
}

func runReadFinish(r *repo.Repo) error {
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
}

func runReadLog(r *repo.Repo) error {
	books, err := r.ListBooks(true)
	if err != nil {
		return err
	}

	input := &repo.ReadInput{}
	if err := forms.LogRead(input, books, r.ListCopies); err != nil {
		return err
	}

	return r.AddRead(input)
}
