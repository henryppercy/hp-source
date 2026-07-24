package cmd

import (
	"fmt"

	"github.com/henryppercy/hp-source/internal/forms"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

func newCopyCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Manage copies of your books",
	}
	cmd.AddCommand(
		newCopyAddCmd(a),
		newCopyEditCmd(a),
	)
	return cmd
}

func newCopyAddCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a copy to an existing book",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCopyAdd(a.repo)
		},
	}
}

func newCopyEditCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit a copy's details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCopyEdit(a.repo)
		},
	}
}

func runCopyAdd(r *repo.Repo) error {
	books, err := r.ListBooks(false)
	if err != nil {
		return err
	}
	if len(books) == 0 {
		fmt.Println("No books to add a copy to.")
		return nil
	}

	var bookID int
	if err := forms.SelectBook(books, &bookID); err != nil {
		return err
	}

	var bookTitle string
	for _, b := range books {
		if b.ID == bookID {
			bookTitle = b.Title
			break
		}
	}

	book, err := r.GetBook(bookID)
	if err != nil {
		return err
	}
	translators, err := r.ListTranslators()
	if err != nil {
		return err
	}

	in := &repo.CopyInput{ShelfStatus: "shelf", SecondHand: true, DateAcquired: today()}
	if err := forms.CopyForm(in, translators, bookTitle, book.OriginalLanguage); err != nil {
		return err
	}

	if err := r.AddCopy(bookID, in); err != nil {
		return err
	}

	fmt.Println("Copy added.")
	return nil
}

func runCopyEdit(r *repo.Repo) error {
	books, err := r.ListBooks(true)
	if err != nil {
		return err
	}
	if len(books) == 0 {
		fmt.Println("No books with copies to edit.")
		return nil
	}

	var bookID, copyID int
	if err := forms.SelectCopy(books, r.ListCopies, &bookID, &copyID); err != nil {
		return err
	}

	detail, err := r.GetCopy(copyID)
	if err != nil {
		return err
	}
	translators, err := r.ListTranslators()
	if err != nil {
		return err
	}
	book, err := r.GetBook(bookID)
	if err != nil {
		return err
	}

	if err := forms.CopyForm(detail, translators, detail.Title, book.OriginalLanguage); err != nil {
		return err
	}

	if err := r.UpdateCopy(copyID, detail); err != nil {
		return err
	}

	fmt.Println("Copy updated.")
	return nil
}
