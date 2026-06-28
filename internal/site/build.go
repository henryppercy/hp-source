package site

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/frostybee/kazari"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/site/templates"
	"github.com/henryppercy/hp-source/internal/text"
	"github.com/yuin/goldmark"
)

// Build renders the site into out from the repo using the embedded assets.
func Build(r *repo.Repo, out string) error {
	return newBuilder(r, embeddedAssets(), out).Build()
}

type builder struct {
	repo   *repo.Repo
	assets fs.FS
	out    string
	engine *kazari.Engine
	md     goldmark.Markdown
}

func newBuilder(r *repo.Repo, assets fs.FS, out string) *builder {
	engine := newCodeEngine()
	return &builder{
		repo:   r,
		assets: assets,
		out:    out,
		engine: engine,
		md:     newMarkdown(engine),
	}
}

// Build wipes the output directory, then renders every page and copies the
// static assets and generated code-block assets into it.
func (b *builder) Build() error {
	if err := os.RemoveAll(b.out); err != nil {
		return fmt.Errorf("failed to clear output directory: %w", err)
	}

	posts, err := b.repo.ListPublishedPosts()
	if err != nil {
		return err
	}
	reads, err := b.repo.ListReads()
	if err != nil {
		return err
	}

	books := bookViews(reads)
	year := time.Now().Year()

	home := templates.HomeView{
		RecentBooks:     recentBooks(books, recentLimit),
		RecentPosts:     recentPosts(posts, recentLimit),
		BooksReadInYear: booksReadInYear(reads, year),
		Year:            year,
	}
	if err := b.render("/", templates.Home(home)); err != nil {
		return err
	}

	if err := b.render("/posts", templates.PostList(templates.PostListView{
		Heading: "Posts",
		Posts:   mainArticles(posts),
	})); err != nil {
		return err
	}

	timeline, err := b.sliceItems(allSlices(posts))
	if err != nil {
		return err
	}
	if err := b.render("/slices", templates.Slices(templates.SliceFeedView{
		Heading: "Slices",
		Intro:   "Get a slice of my life. A personal feed of my thoughts, notes, and updates.",
		Slices:  timeline,
	})); err != nil {
		return err
	}

	if err := b.renderTopics(posts); err != nil {
		return err
	}

	for _, p := range posts {
		view, err := b.postView(p)
		if err != nil {
			return err
		}
		if err := b.render(postURL(p), templates.Post(view)); err != nil {
			return err
		}
	}

	reading := templates.ReadingView{
		CurrentlyReading: booksByStatus(books, "reading"),
		Finished:         booksByStatus(books, "finished"),
	}
	if err := b.render("/reading", templates.Reading(reading)); err != nil {
		return err
	}

	if err := b.render("/404", templates.NotFound()); err != nil {
		return err
	}

	if err := b.copyStatic(); err != nil {
		return err
	}
	return b.writeCodeAssets()
}

// sliceItems renders already-filtered slice posts (newest-first) into timeline
// items, reused by /slices and topic pages.
func (b *builder) sliceItems(slices []repo.Post) ([]templates.SliceItem, error) {
	var items []templates.SliceItem
	for _, p := range slices {
		body, _, err := render(b.md, p.Body)
		if err != nil {
			return nil, fmt.Errorf("slice %s: %w", p.Slug, err)
		}
		items = append(items, templates.SliceItem{
			URL:         postURL(p),
			PublishedAt: parseDate(p.PublishedAt),
			BodyHTML:    body,
			Topics:      topicLinks(p.Topics),
		})
	}
	return items, nil
}

// renderTopics renders /spanish (kept as a special top-level route) plus a
// /topics/{topic} page for every topic present on a published post.
func (b *builder) renderTopics(posts []repo.Post) error {
	spanish, err := b.topicFeed("spanish", posts)
	if err != nil {
		return err
	}
	if err := b.render("/spanish", templates.Spanish(spanish)); err != nil {
		return err
	}

	for _, name := range usedTopics(posts) {
		view, err := b.topicFeed(name, posts)
		if err != nil {
			return err
		}
		if err := b.render("/topics/"+text.Slug(name), templates.Topic(view)); err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) topicFeed(name string, posts []repo.Post) (templates.TopicFeedView, error) {
	slices, err := b.sliceItems(slicesWithTopic(posts, name))
	if err != nil {
		return templates.TopicFeedView{}, err
	}
	return templates.TopicFeedView{
		Heading:  titleCase(name),
		Articles: articlesWithTopic(posts, name),
		Slices:   slices,
	}, nil
}

func (b *builder) postView(p repo.Post) (templates.PostView, error) {
	body, toc, err := render(b.md, p.Body)
	if err != nil {
		return templates.PostView{}, fmt.Errorf("post %s: %w", p.Slug, err)
	}
	return templates.PostView{
		Title:       p.Title,
		Slug:        p.Slug,
		PublishedAt: parseDate(p.PublishedAt),
		UpdatedAt:   parseDate(p.UpdatedAt),
		Headline:    p.Headline,
		BodyHTML:    body,
		TOC:         toc,
		Topics:      topicLinks(p.Topics),
	}, nil
}

// writeCodeAssets emits the kazari stylesheet and progressive-enhancement script
// that style and drive the rendered code blocks.
func (b *builder) writeCodeAssets() error {
	assets := map[string]string{
		filepath.Join("static", "styles", "code.css"): b.engine.CSS(),
		filepath.Join("static", "scripts", "code.js"): b.engine.JS(),
	}
	for rel, content := range assets {
		dest := filepath.Join(b.out, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", dest, err)
		}
		if err := os.WriteFile(dest, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", dest, err)
		}
	}
	return nil
}

func (b *builder) render(urlPath string, c templ.Component) error {
	dest := outputPath(b.out, urlPath)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", dest, err)
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", dest, err)
	}
	defer f.Close()

	if err := c.Render(context.Background(), f); err != nil {
		return fmt.Errorf("failed to render %s: %w", urlPath, err)
	}
	return nil
}

// outputPath maps a URL path to its file on disk:
// "/" -> index.html, "/404" -> 404.html, "/x/y" -> x/y/index.html.
func outputPath(out, urlPath string) string {
	switch urlPath {
	case "/":
		return filepath.Join(out, "index.html")
	case "/404":
		return filepath.Join(out, "404.html")
	default:
		rel := filepath.FromSlash(strings.TrimPrefix(urlPath, "/"))
		return filepath.Join(out, rel, "index.html")
	}
}

func (b *builder) copyStatic() error {
	return fs.WalkDir(b.assets, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := fs.ReadFile(b.assets, path)
		if err != nil {
			return fmt.Errorf("failed to read asset %s: %w", path, err)
		}
		dest := filepath.Join(b.out, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", dest, err)
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", dest, err)
		}
		return nil
	})
}
