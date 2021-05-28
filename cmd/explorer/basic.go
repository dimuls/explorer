//+build basic

package main

import (
	"explorer/hf"
	"explorer/ui"
	"explorer/ui/basic"
)

func init() {
	ui.RegisterUI(basic.FS)
	hf.RegisterJSONValueType("basic", "")
}
