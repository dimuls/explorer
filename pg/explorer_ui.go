package pg

import (
	"embed"
	"explorer/ui"
)

var uiFS = ui.FS

func RegisterUI(fs embed.FS) {
	uiFS = fs
}
