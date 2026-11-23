package templates

// Chrome is the site shell copy shared by every page, the wordmark, the nav
// and the footer. layout.templ reads it directly, like LastBuild. The builder
// sets it from site/chrome.go. Edit the words there, not here.
var Chrome ChromeCopy

// ChromeCopy is the shell copy shared by every page.
type ChromeCopy struct {
	SiteName  string
	FiledFrom string // header label above the location stamp
	Nav       []NavLink
	Footer    FooterCopy
	NotFound  NotFoundCopy
}

// NavLink is one masthead nav entry, a two-digit ordinal, a label and a path.
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

// FooterCopy is the footer wording, from the wordmark and taglines through
// the column headings to the legal line.
type FooterCopy struct {
	Wordmark     string
	TaglineA     string
	TaglineB     string
	BuildHeading string
	SiteHeading  string
	LinksHeading string
	Generator    string // the "generator" value, e.g. "Custom Go SSG"
	JavaScript   string // the "javascript" value, e.g. "0 KB"
	LicenseLabel string
	LicenseURL   string
	Copyright    string
	NoTrackers   string
	Links        []FooterLink
}

// NotFoundCopy is the 404 page wording, the numeral, the line, the nowhere
// stamp and the way back.
type NotFoundCopy struct {
	Code      string
	Message   string
	Place     string
	Coords    string
	BackLabel string
}

// HomeCopy is the homepage prose, carried on HomeView and set from home.go.
// The subheading says who I am and the bio what the site is, so the two do not
// repeat each other.
type HomeCopy struct {
	Kicker      string
	Hero        string
	Subheading  string
	Bio         string
	StreamIntro string
}
