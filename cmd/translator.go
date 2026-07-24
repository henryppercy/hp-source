package cmd

import (
	"fmt"

	"github.com/henryppercy/hp-source/internal/forms"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

func newTranslatorCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "translator",
		Short: "Manage translators",
	}
	cmd.AddCommand(newTranslatorAddCmd(a))
	return cmd
}

func newTranslatorAddCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Add a new translator",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTranslatorAdd(a.repo)
		},
	}
}

func runTranslatorAdd(r *repo.Repo) error {
	var name, sort string
	if err := forms.AddTranslator(&name, &sort); err != nil {
		return err
	}
	if _, err := r.CreateTranslator(name, sort); err != nil {
		return err
	}
	fmt.Println("Translator added.")
	return nil
}
