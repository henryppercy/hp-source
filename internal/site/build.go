package site

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Build(out string) error {
	return newBuilder(embeddedAssets(), out).Build()
}

type builder struct {
	assets fs.FS
	out    string
	funcs  template.FuncMap
}

func newBuilder(assets fs.FS, out string) *builder {
	return &builder{
		assets: assets,
		out:    out,
		funcs:  template.FuncMap{"fmtDate": fmtDate},
	}
}

func fmtDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2 Jan 2006")
}

func (b *builder) Build() error {
	if err := os.RemoveAll(b.out); err != nil {
		return fmt.Errorf("failed to clear output directory: %w", err)
	}

	posts := mockPosts()
	books := mockBooks()

	home := HomeView{
		RecentBooks:     books,
		RecentPosts:     toListItems(posts),
		BooksReadInYear: len(booksByStatus(books, "finished")),
		Year:            time.Now().Year(),
	}
	if err := b.render("/", "home.html", home); err != nil {
		return err
	}

	if err := b.render("/posts", "post_list.html", PostListView{Heading: "Posts", Posts: postListItems(posts, "")}); err != nil {
		return err
	}
	if err := b.render("/spanish", "post_list.html", PostListView{Heading: "Spanish", Posts: postListItems(posts, "spanish")}); err != nil {
		return err
	}
	if err := b.render("/slices", "post_list.html", PostListView{Heading: "Slices", Posts: postListItems(posts, "slice")}); err != nil {
		return err
	}

	for _, p := range posts {
		if err := b.render("/posts/"+p.Slug, "post.html", p); err != nil {
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

	return b.copyStatic()
}

func (b *builder) render(urlPath, page string, data any) error {
	tmpl, err := template.New("layout").Funcs(b.funcs).ParseFS(b.assets, "templates/layout.html", "templates/"+page)
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
