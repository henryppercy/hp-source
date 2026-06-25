package repo

import "fmt"

type Post struct {
	ID        int
	Slug      string
	Title     string
	Type      string
	Headline  string
	Body      string
	PostedAt  string
	CreatedAt string
	UpdatedAt string
}

// returns every post with a posted_at date (i.e. not a draft),
// newest first. The site filters by type and slices for its feeds.
func (r *Repo) ListPublishedPosts() ([]Post, error) {
	rows, err := r.db.Query(
		`SELECT id, slug, title, type, headline, body, posted_at, created_at, updated_at
         FROM post
         WHERE posted_at IS NOT NULL
         ORDER BY posted_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list published posts: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		var headline, postedAt *string
		if err := rows.Scan(
			&p.ID, &p.Slug, &p.Title, &p.Type, &headline, &p.Body,
			&postedAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		if headline != nil {
			p.Headline = *headline
		}
		if postedAt != nil {
			p.PostedAt = *postedAt
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}
