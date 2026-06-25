package site

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/henryppercy/hp-source/internal/repo"
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
	funcs  template.FuncMap
	md     goldmark.Markdown
}

func newBuilder(r *repo.Repo, assets fs.FS, out string) *builder {
	return &builder{
		repo:   r,
		assets: assets,
		out:    out,
		funcs:  template.FuncMap{"fmtDate": fmtDate},
		md:     newMarkdown(),
	}
}

// fmtDate formats a date for display, rendering the zero time as "" so missing
// dates show blank rather than a year-one placeholder.
func fmtDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2 Jan 2006")
}

// Build wipes the output directory, then renders every page and copies the
// static assets and generated chroma stylesheet into it.
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

	home := HomeView{
		RecentBooks:     recentBooks(books, recentLimit),
		RecentPosts:     recentPosts(posts, recentLimit),
		BooksReadInYear: booksReadInYear(reads, year),
		Year:            year,
	}
	if err := b.render("/", "home.html", home); err != nil {
		return err
	}

	if err := b.render("/posts", "post_list.html", PostListView{
		Heading: "Posts",
		Posts:   mainArticles(posts),
	}); err != nil {
		return err
	}

	timeline, err := b.sliceItems(allSlices(posts))
	if err != nil {
		return err
	}
	if err := b.render("/slices", "slices.html", SliceFeedView{
		Heading: "Slices",
		Intro:   "Get a slice of my life. A personal feed of my thoughts, notes, and updates.",
		Slices:  timeline,
	}); err != nil {
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
		if err := b.render(postURL(p), "post.html", view); err != nil {
			return err
		}
	}

	reading := ReadingView{
		CurrentlyReading: booksByStatus(books, "reading"),
		Finished:         booksByStatus(books, "finished"),
	}
	if err := b.render("/reading", "reading.html", reading); err != nil {
		return err
	}

	if err := b.render("/404", "404.html", nil); err != nil {
		return err
	}

	if err := b.copyStatic(); err != nil {
		return err
	}
	return b.writeChromaCSS()
}

// sliceItems renders already-filtered slice posts (newest-first) into timeline
// items, reused by /slices and topic pages.
func (b *builder) sliceItems(slices []repo.Post) ([]SliceItem, error) {
	var items []SliceItem
	for _, p := range slices {
		body, _, err := render(b.md, p.Body)
		if err != nil {
			return nil, fmt.Errorf("slice %s: %w", p.Slug, err)
		}
		items = append(items, SliceItem{
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
	if err := b.render("/spanish", "spanish.html", spanish); err != nil {
		return err
	}

	for _, name := range usedTopics(posts) {
		view, err := b.topicFeed(name, posts)
		if err != nil {
			return err
		}
		if err := b.render("/topics/"+text.Slug(name), "topic.html", view); err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) topicFeed(name string, posts []repo.Post) (TopicFeedView, error) {
	slices, err := b.sliceItems(slicesWithTopic(posts, name))
	if err != nil {
		return TopicFeedView{}, err
	}
	return TopicFeedView{
		Heading:  titleCase(name),
		Articles: articlesWithTopic(posts, name),
		Slices:   slices,
	}, nil
}

func (b *builder) postView(p repo.Post) (PostView, error) {
	body, toc, err := render(b.md, p.Body)
	if err != nil {
		return PostView{}, fmt.Errorf("post %s: %w", p.Slug, err)
	}
	return PostView{
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

func (b *builder) writeChromaCSS() error {
	css, err := chromaCSS()
	if err != nil {
		return err
	}
	dest := filepath.Join(b.out, "static", "chroma.css")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", dest, err)
	}
	if err := os.WriteFile(dest, css, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", dest, err)
	}
	return nil
}

func (b *builder) render(urlPath, page string, data any) error {
	tmpl, err := template.New("layout").Funcs(b.funcs).ParseFS(
		b.assets, "templates/layout.html", "templates/partials.html", "templates/"+page,
	)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", page, err)
	}

	dest := outputPath(b.out, urlPath)
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", dest, err)
	}

	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", dest, err)
	}
	defer f.Close()

	if err := tmpl.ExecuteTemplate(f, "layout", data); err != nil {
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
