package repo

import (
	"database/sql"
	"fmt"
)

type Post struct {
	ID          int
	Slug        string
	Title       string
	Type        string
	Headline    string
	Body        string
	PublishedAt string
	CreatedAt   string
	UpdatedAt   string
}

type PostInput struct {
	ID          int
	Slug        string
	Title       string
	Type        string // "" | "slice" | "spanish"
	Headline    string
	PublishedAt string // "" = draft
}

const postColumns = `id, slug, title, type, headline, body, published_at, created_at, updated_at`

func scanPost(rows *sql.Rows) (Post, error) {
	var p Post
	var headline, publishedAt *string
	if err := rows.Scan(
		&p.ID, &p.Slug, &p.Title, &p.Type, &headline, &p.Body,
		&publishedAt, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return Post{}, fmt.Errorf("failed to scan post: %w", err)
	}
	if headline != nil {
		p.Headline = *headline
	}
	if publishedAt != nil {
		p.PublishedAt = *publishedAt
	}
	return p, nil
}

// ListPublishedPosts returns every post with a published_at date (i.e. not a
// draft), newest first. The site filters by type and slices for its feeds.
func (r *Repo) ListPublishedPosts() ([]Post, error) {
	return r.queryPosts(
		`SELECT ` + postColumns + ` FROM post
         WHERE published_at IS NOT NULL
         ORDER BY published_at DESC`,
	)
}

// ListPosts returns every post, drafts included, most recently changed first,
// for the authoring selection menus.
func (r *Repo) ListPosts() ([]Post, error) {
	return r.queryPosts(`SELECT ` + postColumns + ` FROM post ORDER BY updated_at DESC`)
}

// ListDrafts returns posts without a published_at date, for the publish menu.
func (r *Repo) ListDrafts() ([]Post, error) {
	return r.queryPosts(
		`SELECT ` + postColumns + ` FROM post
         WHERE published_at IS NULL
         ORDER BY updated_at DESC`,
	)
}

func (r *Repo) queryPosts(query string) ([]Post, error) {
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *Repo) GetPost(id int) (Post, error) {
	rows, err := r.db.Query(`SELECT `+postColumns+` FROM post WHERE id = ?`, id)
	if err != nil {
		return Post{}, fmt.Errorf("failed to get post: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return Post{}, fmt.Errorf("failed to get post: %w", err)
		}
		return Post{}, fmt.Errorf("post %d not found", id)
	}
	return scanPost(rows)
}

// CreatePost inserts a post with an empty body and returns its new id. The body
// is written separately via UpdatePostBody.
func (r *Repo) CreatePost(in *PostInput) (int, error) {
	result, err := r.db.Exec(
		`INSERT INTO post (slug, title, type, headline, published_at)
         VALUES (?, ?, ?, ?, ?)`,
		in.Slug, in.Title, in.Type, nullable(in.Headline), nullable(in.PublishedAt),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create post %q: %w", in.Slug, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get post id: %w", err)
	}
	return int(id), nil
}

func (r *Repo) UpdatePost(in *PostInput) error {
	_, err := r.db.Exec(
		`UPDATE post SET slug = ?, title = ?, type = ?, headline = ?, published_at = ?
         WHERE id = ?`,
		in.Slug, in.Title, in.Type, nullable(in.Headline), nullable(in.PublishedAt), in.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}
	return nil
}

func (r *Repo) UpdatePostBody(id int, body string) error {
	if _, err := r.db.Exec(`UPDATE post SET body = ? WHERE id = ?`, body, id); err != nil {
		return fmt.Errorf("failed to update post body: %w", err)
	}
	return nil
}

func (r *Repo) PublishPost(id int, publishedAt string) error {
	if _, err := r.db.Exec(`UPDATE post SET published_at = ? WHERE id = ?`, publishedAt, id); err != nil {
		return fmt.Errorf("failed to publish post: %w", err)
	}
	return nil
}

func (r *Repo) DeletePost(id int) error {
	if _, err := r.db.Exec(`DELETE FROM post WHERE id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	return nil
}
