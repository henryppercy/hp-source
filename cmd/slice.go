package cmd

import (
	"fmt"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/henryppercy/hp-source/internal/forms"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

func newSliceCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slice",
		Short: "Manage site slices (microblog notes)",
	}
	cmd.AddCommand(
		newSliceAddCmd(a),
		newSliceEditCmd(a),
		newSliceWriteCmd(a),
		newSlicePublishCmd(a),
		newSliceDeleteCmd(a),
	)
	return cmd
}

func newSliceAddCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Create a slice (metadata only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSliceAdd(a.repo)
		},
	}
}

func newSliceEditCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit a slice's metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSliceEdit(a.repo)
		},
	}
}

func newSliceWriteCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "write",
		Short: "Write a slice's body in $EDITOR",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSliceWrite(a.repo)
		},
	}
}

func newSlicePublishCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "publish",
		Short: "Publish a draft slice",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSlicePublish(a.repo)
		},
	}
}

func newSliceDeleteCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a slice",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSliceDelete(a.repo)
		},
	}
}

func runSliceAdd(r *repo.Repo) error {
	input := &repo.PostInput{Type: "slice"}
	if err := forms.AddSlice(input); err != nil {
		return err
	}

	if _, err := r.CreatePost(input); err != nil {
		return err
	}

	fmt.Printf("created slice '%s'. run `hp slice write` to add the body.\n", input.Slug)
	return nil
}

func runSliceEdit(r *repo.Repo) error {
	slice, err := pickPost(r.ListSlices, forms.SelectSlice)
	if err != nil {
		return err
	}
	if slice == nil {
		fmt.Println("No slices yet.")
		return nil
	}

	input := &repo.PostInput{
		ID:          slice.ID,
		Slug:        slice.Slug,
		Type:        slice.Type,
		PublishedAt: slice.PublishedAt,
	}
	if err := forms.EditSlice(input); err != nil {
		return err
	}

	if err := r.UpdatePost(input); err != nil {
		return err
	}

	fmt.Printf("updated '%s'.\n", input.Slug)
	return nil
}

func runSliceWrite(r *repo.Repo) error {
	slice, err := pickPost(r.ListSlices, forms.SelectSlice)
	if err != nil {
		return err
	}
	if slice == nil {
		fmt.Println("No slices yet.")
		return nil
	}

	body, err := editBody(slice.Body)
	if err != nil {
		return err
	}
	if body == slice.Body {
		fmt.Println("no changes.")
		return nil
	}

	if err := r.UpdatePostBody(slice.ID, body); err != nil {
		return err
	}

	fmt.Printf("saved body of '%s'.\n", slice.Slug)
	return nil
}

func runSlicePublish(r *repo.Repo) error {
	slice, err := pickPost(r.ListSliceDrafts, forms.SelectSlice)
	if err != nil {
		return err
	}
	if slice == nil {
		fmt.Println("No draft slices to publish.")
		return nil
	}

	date := ""
	if err := forms.PublishPost(slice.Slug, &date); err != nil {
		return err
	}

	if err := r.PublishPost(slice.ID, date); err != nil {
		return err
	}

	fmt.Printf("published '%s' (%s).\n", slice.Slug, date)
	return nil
}

func runSliceDelete(r *repo.Repo) error {
	slice, err := pickPost(r.ListSlices, forms.SelectSlice)
	if err != nil {
		return err
	}
	if slice == nil {
		fmt.Println("No slices yet.")
		return nil
	}

	var confirm bool
	err = huh.NewConfirm().
		WithButtonAlignment(lipgloss.Left).
		Title(fmt.Sprintf("Delete '%s'?", slice.Slug)).
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

	if err := r.DeletePost(slice.ID); err != nil {
		return err
	}

	fmt.Printf("deleted '%s'.\n", slice.Slug)
	return nil
}
