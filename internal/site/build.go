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
	shelf, err := b.repo.ListShelf()
	if err != nil {
		return err
	}
	spanishLog, err := b.repo.ListSpanishLog()
	if err != nil {
		return err
	}

	if err := b.loadLocations(); err != nil {
		return err
	}
	// A recorded build (hp site build) has already set LastBuild; dev serving
	// and watch rebuilds fall back to live info filed from home.
	if templates.LastBuild.Date.IsZero() {
		templates.LastBuild = liveBuildInfo()
	}

	now := time.Now()
	year := now.Year()

	timeline, err := b.sliceItems(allSlices(posts))
	if err != nil {
		return err
	}

	home := homeView(posts, reads, timeline, spanishLog, now)
	if err := b.render("/", templates.Home(home)); err != nil {
		return err
	}

	if err := b.render("/posts", templates.PostList(templates.PostListView{
		Heading: "Posts",
		Posts:   mainArticles(posts),
	})); err != nil {
		return err
	}

	if err := b.render("/slices", templates.Slices(templates.SliceFeedView{
		Heading: "Slices",
		Intro:   "Get a slice of my life. A personal feed of my thoughts, notes, and updates.",
		Slices:  timeline,
	})); err != nil {
		return err
	}

	if err := b.renderTopics(posts, spanishLog); err != nil {
		return err
	}

	for _, p := range posts {
		if p.Kind == "slice" {
			item, err := b.sliceItem(p)
			if err != nil {
				return err
			}
			if err := b.render(postURL(p), templates.Slice(item)); err != nil {
				return err
			}
			continue
		}
		view, err := b.postView(p)
		if err != nil {
			return err
		}
		if err := b.render(postURL(p), templates.Post(view)); err != nil {
			return err
		}
	}

	if err := b.render("/reading", templates.Reading(readingHub(reads, shelf, year))); err != nil {
		return err
	}
	for _, p := range readingYearPages(reads, year) {
		if err := b.render(fmt.Sprintf("/reading/%d", p.Year), templates.ReadingYear(p.View)); err != nil {
			return err
		}
	}
	if err := b.render("/reading/shelf", templates.ReadingShelf(readingShelf(shelf))); err != nil {
		return err
	}

	if err := b.render("/design", templates.Design()); err != nil {
		return err
	}

	if err := b.render("/404", templates.NotFound()); err != nil {
		return err
	}

	if err := b.copyStatic(); err != nil {
		return err
	}
	return b.writeCodeCSS()
}

// sliceItem renders one slice post into a timeline item, shared by the feed and
// the slice permalink page.
func (b *builder) sliceItem(p repo.Post) (templates.SliceItem, error) {
	body, _, err := render(b.md, p.Body)
	if err != nil {
		return templates.SliceItem{}, fmt.Errorf("slice %s: %w", p.Slug, err)
	}
	return templates.SliceItem{
		URL:         postURL(p),
		Slug:        p.Slug,
		PublishedAt: parseDate(p.PublishedAt),
		BodyHTML:    body,
		Topics:      topicLinks(p.Topics),
		Location:    locationStamps[p.LocationID],
	}, nil
}

// sliceItems renders already-filtered slice posts (newest-first) into timeline
// items, reused by /slices and topic pages.
func (b *builder) sliceItems(slices []repo.Post) ([]templates.SliceItem, error) {
	var items []templates.SliceItem
	for _, p := range slices {
		item, err := b.sliceItem(p)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// renderTopics renders /spanish (kept as a special top-level route) plus a
// /topics/{topic} page for every topic present on a published post.
func (b *builder) renderTopics(posts []repo.Post, log []repo.SpanishLogEntry) error {
	slices, err := b.sliceItems(slicesWithTopic(posts, "spanish"))
	if err != nil {
		return err
	}
	spanish := spanishView(log, articlesWithTopic(posts, "spanish"), slices, time.Now())
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
		ReadMinutes: text.ReadMinutes(p.Body),
		Words:       text.WordCount(p.Body),
		TOC:         toc,
		Topics:      topicLinks(p.Topics),
		Location:    locationStamps[p.LocationID],
	}, nil
}

// loadLocations caches every place by id into locationStamps for the post and
// slice cards, and sets the shared home location used by the header, footer and
// home nameplate.
func (b *builder) loadLocations() error {
	locations, err := b.repo.ListLocations()
	if err != nil {
		return err
	}
	locationStamps = make(map[int]templates.Place, len(locations))
	for _, l := range locations {
		locationStamps[l.ID] = placeOf(l)
	}

	home, err := b.repo.GetLocationBySlug(HomeSlug)
	if err != nil {
		return err
	}
	templates.HomeLocation = placeOf(*home)
	return nil
}

// writeCodeCSS emits the kazari stylesheet that styles the rendered code blocks.
func (b *builder) writeCodeCSS() error {
	dest := filepath.Join(b.out, "static", "styles", "code.css")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", dest, err)
	}
	if err := os.WriteFile(dest, []byte(b.engine.CSS()), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", dest, err)
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
