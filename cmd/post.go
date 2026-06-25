package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/henryppercy/hp-source/internal/forms"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/spf13/cobra"
)

func newPostCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post",
		Short: "Manage site posts",
	}
	cmd.AddCommand(
		newPostAddCmd(a),
		newPostEditCmd(a),
		newPostWriteCmd(a),
		newPostPublishCmd(a),
		newPostDeleteCmd(a),
	)
	return cmd
}

func newPostAddCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Create a post (metadata only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPostAdd(a.repo)
		},
	}
}

func newPostEditCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit a post's metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPostEdit(a.repo)
		},
	}
}

func newPostWriteCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "write",
		Short: "Write a post's body in $EDITOR",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPostWrite(a.repo)
		},
	}
}

func newPostPublishCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "publish",
		Short: "Publish a draft",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPostPublish(a.repo)
		},
	}
}

func newPostDeleteCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a post",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPostDelete(a.repo)
		},
	}
}

func runPostAdd(r *repo.Repo) error {
	input := &repo.PostInput{}
	if err := forms.AddPost(input); err != nil {
		return err
	}

	if _, err := r.CreatePost(input); err != nil {
		return err
	}

	if input.PublishedAt == "" {
		fmt.Printf("created draft '%s'. run `hp post write` to edit the body.\n", input.Slug)
	} else {
		fmt.Printf("created post '%s'. run `hp post write` to edit the body.\n", input.Slug)
	}
	return nil
}

func runPostEdit(r *repo.Repo) error {
	post, err := selectPost(r.ListPosts)
	if err != nil {
		return err
	}
	if post == nil {
		fmt.Println("No posts yet.")
		return nil
	}

	input := &repo.PostInput{
		ID:          post.ID,
		Slug:        post.Slug,
		Title:       post.Title,
		Type:        post.Type,
		Headline:    post.Headline,
		PublishedAt: post.PublishedAt,
	}
	if err := forms.EditPost(input); err != nil {
		return err
	}

	if err := r.UpdatePost(input); err != nil {
		return err
	}

	fmt.Printf("updated '%s'.\n", input.Slug)
	return nil
}

func runPostWrite(r *repo.Repo) error {
	post, err := selectPost(r.ListPosts)
	if err != nil {
		return err
	}
	if post == nil {
		fmt.Println("No posts yet.")
		return nil
	}

	body, err := editBody(post.Body)
	if err != nil {
		return err
	}
	if body == post.Body {
		fmt.Println("no changes.")
		return nil
	}

	if err := r.UpdatePostBody(post.ID, body); err != nil {
		return err
	}

	fmt.Printf("saved body of '%s'.\n", post.Slug)
	return nil
}

func runPostPublish(r *repo.Repo) error {
	post, err := selectPost(r.ListDrafts)
	if err != nil {
		return err
	}
	if post == nil {
		fmt.Println("No drafts to publish.")
		return nil
	}

	date := ""
	if err := forms.PublishPost(post.Title, &date); err != nil {
		return err
	}

	if err := r.PublishPost(post.ID, date); err != nil {
		return err
	}

	fmt.Printf("published '%s' (%s).\n", post.Slug, date)
	return nil
}

func runPostDelete(r *repo.Repo) error {
	post, err := selectPost(r.ListPosts)
	if err != nil {
		return err
	}
	if post == nil {
		fmt.Println("No posts yet.")
		return nil
	}

	var confirm bool
	err = huh.NewConfirm().
		WithButtonAlignment(lipgloss.Left).
		Title(fmt.Sprintf("Delete '%s'?", post.Title)).
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

	if err := r.DeletePost(post.ID); err != nil {
		return err
	}

	fmt.Printf("deleted '%s'.\n", post.Slug)
	return nil
}

// selectPost lists posts via list, returns the chosen one, or nil when there
// are none to act on.
func selectPost(list func() ([]repo.Post, error)) (*repo.Post, error) {
	posts, err := list()
	if err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return nil, nil
	}

	id, err := forms.SelectPost(posts)
	if err != nil {
		return nil, err
	}
	for i := range posts {
		if posts[i].ID == id {
			return &posts[i], nil
		}
	}
	return nil, nil
}

// editBody opens initial in $EDITOR (a temp *.md file) and returns the saved
// contents.
func editBody(initial string) (string, error) {
	f, err := os.CreateTemp("", "hp-post-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(initial); err != nil {
		f.Close()
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	parts := strings.Fields(editor)
	c := exec.Command(parts[0], append(parts[1:], f.Name())...)
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("editor failed: %w", err)
	}

	data, err := os.ReadFile(f.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read temp file: %w", err)
	}
	return string(data), nil
}
