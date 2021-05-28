package ui

import "embed"

var FS embed.FS

func RegisterUI(fs embed.FS) {
	FS = fs
}
