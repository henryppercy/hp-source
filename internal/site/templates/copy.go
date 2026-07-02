package templates

// Chrome is the site shell copy shared by every page: the wordmark, the nav and
// the footer. layout.templ reads it directly, like LastBuild; the builder sets
// it from site/chrome.go. Edit the words there, not here.
var Chrome ChromeCopy

type ChromeCopy struct {
	SiteName  string
	FiledFrom string // header label above the location stamp
	Nav       []NavLink
	Footer    FooterCopy
	NotFound  NotFoundCopy
}

// NavLink is one masthead nav entry: a two-digit ordinal, a label and a path.
type NavLink struct {
	Num   string
	Label string
	Href  string
}

// FooterLink is one external link in the footer's Links column.
type FooterLink struct {
	Label string
	Href  string
}

// FooterCopy is the colophon: the wordmark, the two taglines, the column
// headings, the two build values that are prose, and the legal line.
type FooterCopy struct {
	Wordmark     string
	TaglineA     string
	TaglineB     string
	BuildHeading string
	SiteHeading  string
	LinksHeading string
	Generator    string // the "generator" value, e.g. "Custom Go SSG"
	JavaScript   string // the "javascript" value, e.g. "0 KB; none"
	LicenseLabel string
	LicenseURL   string
	Copyright    string
	NoTrackers   string
	Links        []FooterLink
}

// NotFoundCopy is the 404 page: the numeral, the line, the nowhere stamp and the
// way back.
type NotFoundCopy struct {
	Code      string
	Message   string
	Place     string
	Coords    string
	BackLabel string
}

// HomeCopy is the frontispiece prose, carried on HomeView and set from home.go.
// The standfirst says who I am, the bio what the site is, so the two do not
// repeat each other.
type HomeCopy struct {
	Kicker        string
	Hero          string
	Standfirst    string
	Bio           string
	StreamIntro   string
	DispatchLeft  string // the dispatch strip's left kicker
	DispatchRight string // its right kicker
}
