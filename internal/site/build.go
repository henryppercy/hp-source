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
	"github.com/charmbracelet/log"
	"github.com/frostybee/kazari"
	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/site/templates"
	"github.com/henryppercy/hp-source/internal/text"
	"github.com/yuin/goldmark"
)

// imageDir is the path of the image tree within the embedded assets FS.
const imageDir = "static/images"

// Build renders the site into out from the repo. When images is
// false the local image tree is left out of the output, as it is
// served from the CDN instead.
func Build(r *repo.Repo, out string, images bool) error {
	return newBuilder(r, embeddedAssets(), out, images, false).Build()
}

type builder struct {
	repo   *repo.Repo
	assets fs.FS
	out    string
	images bool
	drafts bool
	engine *kazari.Engine
	md     goldmark.Markdown
	pages  int // pages rendered by the current Build, reset each run
}

func newBuilder(r *repo.Repo, assets fs.FS, out string, images, drafts bool) *builder {
	engine := newCodeEngine()
	return &builder{
		repo:   r,
		assets: assets,
		out:    out,
		images: images,
		drafts: drafts,
		engine: engine,
		md:     newMarkdown(engine),
	}
}

// Build wipes the output directory, then renders every page and copies the
// static assets and generated code-block assets into it. It logs a summary of
// the pages written and how long the run took.
func (b *builder) Build() error {
	start := time.Now()
	b.pages = 0
	if err := os.RemoveAll(b.out); err != nil {
		return fmt.Errorf("failed to clear output directory: %w", err)
	}
	templates.Chrome = chrome
	loadImageConfig()

	posts, err := b.listPosts()
	if err != nil {
		return err
	}
	dateDrafts(posts)
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

	bookPages, err := b.loadBookPages()
	if err != nil {
		return err
	}

	authorPages, err := b.loadAuthorPages()
	if err != nil {
		return err
	}

	genrePages, err := b.loadGenrePages()
	if err != nil {
		return err
	}

	kindPages, err := b.loadKindPages()
	if err != nil {
		return err
	}

	tagPages, err := b.loadTagPages()
	if err != nil {
		return err
	}

	if err := b.loadLocations(); err != nil {
		return err
	}

	if templates.LastBuild.Date.IsZero() {
		templates.LastBuild = liveBuildInfo()
	}

	now := time.Now()
	year := now.Year()

	timeline, err := b.sliceItems(allSlices(posts))
	if err != nil {
		return err
	}

	stream := streamEntries(posts, reads, timeline, spanishLog)

	home := homeView(posts, reads, timeline, spanishLog, stream, now)
	if err := b.render("/", templates.Home(home)); err != nil {
		return err
	}

	if err := b.render("/posts", templates.PostList(templates.PostListView{
		Heading:    postsHeading,
		Subheading: postsSubheading,
		Empty:      feedEmpty,
		Posts:      mainArticles(posts),
	})); err != nil {
		return err
	}

	if err := b.render("/slices", templates.Slices(templates.SliceFeedView{
		Heading:    slicesHeading,
		Subheading: slicesSubheading,
		Empty:      feedEmpty,
		Slices:     timeline,
	})); err != nil {
		return err
	}

	if err := b.render("/everything", templates.Everything(everythingFeed(stream))); err != nil {
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

	if err := b.renderBookPages(bookPages); err != nil {
		return err
	}

	if err := b.renderAuthorPages(authorPages); err != nil {
		return err
	}

	if err := b.renderGenrePages(genrePages); err != nil {
		return err
	}

	if err := b.renderKindPages(kindPages); err != nil {
		return err
	}

	if err := b.renderTagPages(tagPages); err != nil {
		return err
	}

	if err := b.render("/404", templates.NotFound()); err != nil {
		return err
	}

	if err := b.copyStatic(); err != nil {
		return err
	}
	if err := b.writeCodeCSS(); err != nil {
		return err
	}
	log.Info("built", "pages", b.pages, "out", b.out, "in", time.Since(start).Round(time.Millisecond))
	return nil
}

func loadImageConfig() {
	if p := os.Getenv("HP_IMAGE_PATTERN"); p != "" {
		imagePattern = p
	}
}

func (b *builder) listPosts() ([]repo.Post, error) {
	if b.drafts {
		return b.repo.ListPostsWithDrafts()
	}
	return b.repo.ListPublishedPosts()
}

// dateDrafts gives draft posts their last edit date as the display date, so
// feeds, datelines and permalinks render them sensibly before publishing.
func dateDrafts(posts []repo.Post) {
	for i := range posts {
		if posts[i].PublishedAt == "" {
			posts[i].PublishedAt = posts[i].UpdatedAt
		}
	}
}

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

func (b *builder) renderTopics(posts []repo.Post, log []repo.SpanishLogEntry) error {
	spanish := spanishView(log, articlesWithTopic(posts, "spanish"), time.Now())
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
		Empty:    feedEmpty,
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
	b.pages++
	return nil
}

// outputPath maps a URL path to its file on disk, so "/" is index.html,
// "/404" is 404.html, and "/x/y" is x/y/index.html.
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

		// Images being false means the page loads images from the CDN,
		// therefore local files are not needed in the output.
		if !b.images && (path == imageDir || strings.HasPrefix(path, imageDir+"/")) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
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
