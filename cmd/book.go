package cmd

import (
	"fmt"

	"github.com/henryppercy/hp-source/internal/forms"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

func newBookCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "book",
		Short: "Manage your book collection",
	}
	cmd.AddCommand(
		newBookAddCmd(a),
		newBookEditCmd(a),
	)
	return cmd
}

func newBookAddCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a new book",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBookAdd(a.repo)
		},
	}
}

func newBookEditCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit a book's work details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBookEdit(a.repo)
		},
	}
}

func runBookAdd(r *repo.Repo) error {
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
	translators, err := r.ListTranslators()
	if err != nil {
		return err
	}

	input := &repo.BookInput{}
	addCopy, err := forms.AddBook(input, genres, authors, tags, series)
	if err != nil {
		return err
	}

	var firstCopy *repo.CopyInput
	if addCopy {
		firstCopy = &repo.CopyInput{ShelfStatus: "shelf", SecondHand: true}
		if err := forms.CopyForm(firstCopy, translators, input.Title, input.OriginalLanguage); err != nil {
			return err
		}
	}

	return r.AddBook(input, firstCopy)
}

func runBookEdit(r *repo.Repo) error {
	books, err := r.ListBooks(false)
	if err != nil {
		return err
	}
	if len(books) == 0 {
		fmt.Println("No books to edit.")
		return nil
	}

	var bookID int
	if err := forms.SelectBook(books, &bookID); err != nil {
		return err
	}

	detail, err := r.GetBook(bookID)
	if err != nil {
		return err
	}

	genres, err := r.ListGenres()
	if err != nil {
		return err
	}
	if err := forms.EditBook(detail, genres); err != nil {
		return err
	}

	if err := r.UpdateBook(bookID, detail); err != nil {
		return err
	}

	fmt.Println("Book updated.")
	return nil
}
