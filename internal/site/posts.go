package site

import (
	"sort"

	"github.com/henryppercy/hp-source/internal/repo"
	"github.com/henryppercy/hp-source/internal/site/templates"
	"github.com/henryppercy/hp-source/internal/text"
)

// Section copy for the post, slice and topic listings. Edit the wording here.
const (
	postsHeading    = "Posts"
	postsStandfirst = "Longer form writing; about anything I feel like writing about."

	slicesHeading    = "Slices"
	slicesStandfirst = "Get a slice of my life. A personal feed of my thoughts, notes, and updates."

	feedEmpty = "Nothing here yet."
)

func toListItem(p repo.Post) templates.PostListItem {
	return templates.PostListItem{
		Title:       p.Title,
		Slug:        p.Slug,
		URL:         postURL(p),
		PublishedAt: parseDate(p.PublishedAt),
		Headline:    p.Headline,
		ReadMinutes: text.ReadMinutes(p.Body),
		Topics:      topicLinks(p.Topics),
		Location:    locationStamps[p.LocationID],
	}
}

// postURL is a post's canonical page path. Slices live under /slices, articles
// under /posts.
func postURL(p repo.Post) string {
	if p.Kind == "slice" {
		return "/slices/" + p.Slug
	}
	return "/posts/" + p.Slug
}

// topicLinks maps a post's topics to display links to their feeds.
func topicLinks(topics []repo.Topic) []templates.TopicLink {
	links := make([]templates.TopicLink, len(topics))
	for i, t := range topics {
		links[i] = templates.TopicLink{Name: t.Name, URL: "/topics/" + text.Slug(t.Name)}
	}
	return links
}

func hasTopic(p repo.Post, name string) bool {
	for _, t := range p.Topics {
		if t.Name == name {
			return true
		}
	}
	return false
}

// articleItems maps the articles matching keep to list items.
func articleItems(posts []repo.Post, keep func(repo.Post) bool) []templates.PostListItem {
	var items []templates.PostListItem
	for _, p := range posts {
		if p.Kind == "article" && keep(p) {
			items = append(items, toListItem(p))
		}
	}
	return items
}

// mainArticles are the articles for /posts and the home stream: every article
// except the Spanish learning log (posts whose only topic is spanish, which live
// on /spanish). Posts that touch spanish among other topics still appear.
func mainArticles(posts []repo.Post) []templates.PostListItem {
	return articleItems(posts, func(p repo.Post) bool {
		return !onlySpanish(p)
	})
}

// onlySpanish reports whether spanish is the post's sole topic.
func onlySpanish(p repo.Post) bool {
	return len(p.Topics) == 1 && p.Topics[0].Name == "spanish"
}

// articlesWithTopic maps the articles carrying the named topic to list items.
func articlesWithTopic(posts []repo.Post, name string) []templates.PostListItem {
	return articleItems(posts, func(p repo.Post) bool {
		return hasTopic(p, name)
	})
}

// slicesWithTopic returns the slice posts carrying the named topic.
func slicesWithTopic(posts []repo.Post, name string) []repo.Post {
	var out []repo.Post
	for _, p := range posts {
		if p.Kind == "slice" && hasTopic(p, name) {
			out = append(out, p)
		}
	}
	return out
}

// allSlices returns the slice posts, preserving order.
func allSlices(posts []repo.Post) []repo.Post {
	var out []repo.Post
	for _, p := range posts {
		if p.Kind == "slice" {
			out = append(out, p)
		}
	}
	return out
}

// usedTopics returns the distinct topic names present on the posts, sorted.
func usedTopics(posts []repo.Post) []string {
	seen := map[string]bool{}
	var names []string
	for _, p := range posts {
		for _, t := range p.Topics {
			if !seen[t.Name] {
				seen[t.Name] = true
				names = append(names, t.Name)
			}
		}
	}
	sort.Strings(names)
	return names
}
