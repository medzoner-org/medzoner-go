package resources

import (
	"embed"
	"io/fs"
)

//go:embed editorial/*.md
var editorialFiles embed.FS

// EditorialFiles exposes a read-only, deployment-independent content source.
func EditorialFiles() (fs.FS, error) { return fs.Sub(editorialFiles, "editorial") }
