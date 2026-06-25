package site

import (
	"embed"
	"io/fs"
)

//go:generate ../../bin/tailwindcss -i styles/input.css -o static/styles/app.css --minify

//go:embed templates/*.html static
var embeddedFS embed.FS

func embeddedAssets() fs.FS { return embeddedFS }
