package site

import (
	"embed"
	"io/fs"
)

//go:generate go tool templ generate
//go:generate ../../bin/tailwindcss -i styles/input.css -o static/styles/app.css --minify

//go:embed static
var embeddedFS embed.FS

func embeddedAssets() fs.FS { return embeddedFS }
