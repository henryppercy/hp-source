package site

import "github.com/henryppercy/hp-source/internal/site/templates"

// chrome is the copy shared by every page's shell: the wordmark, the masthead
// nav, and the footer colophon. Edit the site's chrome wording here; the builder
// pushes it into the templates layer (see Build).
var chrome = templates.ChromeCopy{
	SiteName:  "Henry Percy",
	FiledFrom: "filed from",
	Nav: []templates.NavLink{
		{Num: "01", Label: "Home", Href: "/"},
		{Num: "02", Label: "Posts", Href: "/posts"},
		{Num: "03", Label: "Slices", Href: "/slices"},
		{Num: "04", Label: "Reading", Href: "/reading"},
		{Num: "05", Label: "Spanish", Href: "/spanish"},
	},
	Footer: templates.FooterCopy{
		Wordmark:     "Henry Percy",
		TaglineA:     "Statically generated from my life.",
		TaglineB:     "Writing by me; photos taken by me; data collected on me, by me.",
		BuildHeading: "The Build",
		SiteHeading:  "Site",
		LinksHeading: "Links",
		Generator:    "HP Source",
		JavaScript:   "0 KB; none",
		LicenseLabel: "CC BY-NC-SA 4.0",
		LicenseURL:   "https://creativecommons.org/licenses/by-nc-sa/4.0/",
		Copyright:    "2023-26 © Henry Percy",
		NoTrackers:   "No trackers; No JavaScript",
		Links: []templates.FooterLink{
			{Label: "GitHub", Href: "https://github.com/henryppercy"},
			{Label: "LinkedIn", Href: "https://www.linkedin.com/in/henry-b-percy"},
			{Label: "RSS feed", Href: "#"},
			{Label: "Email", Href: "#"},
		},
	},
	NotFound: templates.NotFoundCopy{
		Code:      "404",
		Message:   "Nothing is filed at this address.",
		Place:     "Nowhere",
		Coords:    "--.--°N --.--°W",
		BackLabel: "Back on track",
	},
}
