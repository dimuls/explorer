//+build basic

package ui

import (
	"embed"
)

const Prefix = "dist_basic"

//go:embed dist_basic
var FS embed.FS
