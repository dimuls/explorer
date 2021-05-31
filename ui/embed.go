package ui

import (
	"embed"
)

const Prefix = "dist"

//go:embed dist
var FS embed.FS
