package site

import (
	"embed"
	"io/fs"
)

//go:embed templates/*.html static
var embeddedFS embed.FS

func embeddedAssets() fs.FS { return embeddedFS }
