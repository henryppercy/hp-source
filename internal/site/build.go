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

	feeds := []struct{ path, heading, typ string }{
		{"/posts", "Posts", ""},
		{"/spanish", "Spanish", "spanish"},
		{"/slices", "Slices", "slice"},
	}
	for _, f := range feeds {
		view := PostListView{Heading: f.heading, Posts: listItemsByType(posts, f.typ)}
		if err := b.render(f.path, "post_list.html", view); err != nil {
			return err
		}
	}

	for _, p := range posts {
		view, err := b.postView(p)
		if err != nil {
			return err
		}
		if err := b.render("/posts/"+p.Slug, "post.html", view); err != nil {
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

func (b *builder) postView(p repo.Post) (PostView, error) {
	body, toc, err := render(b.md, p.Body)
	if err != nil {
		return PostView{}, fmt.Errorf("post %s: %w", p.Slug, err)
	}
	return PostView{
		Title:       p.Title,
		Slug:        p.Slug,
		Type:        p.Type,
		PublishedAt: parseDate(p.PublishedAt),
		UpdatedAt:   parseDate(p.UpdatedAt),
		Headline:    p.Headline,
		BodyHTML:    body,
		TOC:         toc,
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
		b.assets, "templates/layout.html", "templates/"+page,
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
