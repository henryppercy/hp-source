package repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/henryppercy/hp-source/internal/text"
)

type Post struct {
	ID          int
	Slug        string
	Title       string
	Kind        string // article | slice
	Headline    string
	Body        string
	PublishedAt string
	CreatedAt   string
	UpdatedAt   string
	LocationID  int // 0 = none
	Topics      []Topic
}

type PostInput struct {
	ID          int
	Slug        string
	Title       string
	Kind        string // article | slice
	Headline    string
	PublishedAt string // "" = draft
	LocationID  int    // 0 = none
	TopicIDs    []int
}

const postColumns = `id, slug, title, kind, headline, body, published_at, created_at, updated_at, location_id`

func scanPost(rows *sql.Rows) (Post, error) {
	var p Post
	var headline, publishedAt *string
	var locationID *int
	if err := rows.Scan(
		&p.ID, &p.Slug, &p.Title, &p.Kind, &headline, &p.Body,
		&publishedAt, &p.CreatedAt, &p.UpdatedAt, &locationID,
	); err != nil {
		return Post{}, fmt.Errorf("failed to scan post: %w", err)
	}
	if headline != nil {
		p.Headline = *headline
	}
	if publishedAt != nil {
		p.PublishedAt = *publishedAt
	}
	if locationID != nil {
		p.LocationID = *locationID
	}
	return p, nil
}

// ListPublishedPosts returns every post with a published_at date (i.e. not a
// draft), newest first. The site filters by kind and topic for its feeds.
func (r *Repo) ListPublishedPosts() ([]Post, error) {
	return r.queryPosts(
		`SELECT ` + postColumns + ` FROM post
         WHERE published_at IS NOT NULL
         ORDER BY published_at DESC`,
	)
}

// ListArticles returns articles (not slices), most recently changed first, for
// the hp post selection menus.
func (r *Repo) ListArticles() ([]Post, error) {
	return r.queryPosts(
		`SELECT ` + postColumns + ` FROM post
         WHERE kind = 'article'
         ORDER BY updated_at DESC`,
	)
}

// ListArticleDrafts returns unpublished articles, for hp post publish.
func (r *Repo) ListArticleDrafts() ([]Post, error) {
	return r.queryPosts(
		`SELECT ` + postColumns + ` FROM post
         WHERE kind = 'article' AND published_at IS NULL
         ORDER BY updated_at DESC`,
	)
}

// ListSlices returns slices, most recently changed first, for the hp slice menus.
func (r *Repo) ListSlices() ([]Post, error) {
	return r.queryPosts(
		`SELECT ` + postColumns + ` FROM post
         WHERE kind = 'slice'
         ORDER BY updated_at DESC`,
	)
}

// ListSliceDrafts returns unpublished slices, for hp slice publish.
func (r *Repo) ListSliceDrafts() ([]Post, error) {
	return r.queryPosts(
		`SELECT ` + postColumns + ` FROM post
         WHERE kind = 'slice' AND published_at IS NULL
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.attachTopics(posts); err != nil {
		return nil, err
	}
	return posts, nil
}

// attachTopics fills each post's Topics in one query over post_topic.
func (r *Repo) attachTopics(posts []Post) error {
	if len(posts) == 0 {
		return nil
	}
	rows, err := r.db.Query(
		`SELECT pt.post_id, t.id, t.name
         FROM post_topic pt JOIN topic t ON t.id = pt.topic_id
         ORDER BY t.name`,
	)
	if err != nil {
		return fmt.Errorf("failed to load topics: %w", err)
	}
	defer rows.Close()

	byPost := map[int][]Topic{}
	for rows.Next() {
		var postID int
		var t Topic
		if err := rows.Scan(&postID, &t.ID, &t.Name); err != nil {
			return fmt.Errorf("failed to scan post topic: %w", err)
		}
		byPost[postID] = append(byPost[postID], t)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range posts {
		posts[i].Topics = byPost[posts[i].ID]
	}
	return nil
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
	p, err := scanPost(rows)
	if err != nil {
		return Post{}, err
	}
	rows.Close()

	posts := []Post{p}
	if err := r.attachTopics(posts); err != nil {
		return Post{}, err
	}
	return posts[0], nil
}

// CreatePost inserts a post with an empty body and returns its new id. The body
// is written separately via UpdatePostBody. The slug is derived (override, else
// title, else a unique date slug for titleless slices); in.Slug is updated to
// the slug actually used.
func (r *Repo) CreatePost(in *PostInput) (int, error) {
	in.Slug = slugFor(in)
	if in.Slug == "" {
		slug, err := r.uniqueDateSlug(time.Now().Format("2006-01-02"))
		if err != nil {
			return 0, err
		}
		in.Slug = slug
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`INSERT INTO post (slug, title, kind, headline, published_at, location_id)
         VALUES (?, ?, ?, ?, ?, ?)`,
		in.Slug, in.Title, in.Kind, nullable(in.Headline), nullable(in.PublishedAt), nullableInt(in.LocationID),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create post %q: %w", in.Slug, err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get post id: %w", err)
	}

	for _, topicID := range in.TopicIDs {
		if err := linkPostTopic(tx, int(id), topicID); err != nil {
			return 0, fmt.Errorf("failed to link topic: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit: %w", err)
	}
	return int(id), nil
}

// slugFor normalises the chosen slug: an explicit override wins, else it is
// derived from the title, else empty (the signal to use a date slug).
func slugFor(in *PostInput) string {
	if s := text.Slug(in.Slug); s != "" {
		return s
	}
	return text.Slug(in.Title)
}

// uniqueDateSlug returns date, or date-2, date-3, ... for the first one not
// already taken, so multiple slices on the same day get distinct slugs.
func (r *Repo) uniqueDateSlug(date string) (string, error) {
	for slug, n := date, 2; ; slug, n = fmt.Sprintf("%s-%d", date, n), n+1 {
		var one int
		err := r.db.QueryRow("SELECT 1 FROM post WHERE slug = ?", slug).Scan(&one)
		if err == sql.ErrNoRows {
			return slug, nil
		}
		if err != nil {
			return "", fmt.Errorf("failed to check slug: %w", err)
		}
	}
}

// UpdatePost updates a post's metadata and replaces its topic links. It does
// not change kind (set once at creation) or the body.
func (r *Repo) UpdatePost(in *PostInput) error {
	in.Slug = slugFor(in)

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE post SET slug = ?, title = ?, headline = ?, published_at = ?, location_id = ?
         WHERE id = ?`,
		in.Slug, in.Title, nullable(in.Headline), nullable(in.PublishedAt), nullableInt(in.LocationID), in.ID,
	); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	if _, err := tx.Exec(`DELETE FROM post_topic WHERE post_id = ?`, in.ID); err != nil {
		return fmt.Errorf("failed to clear topics: %w", err)
	}
	for _, topicID := range in.TopicIDs {
		if err := linkPostTopic(tx, in.ID, topicID); err != nil {
			return fmt.Errorf("failed to link topic: %w", err)
		}
	}

	return tx.Commit()
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
