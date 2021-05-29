package main

import (
	"explorer"
	"explorer/hf"
)

func main() {
	hf.RegisterJSONValueType("basic", "")
	explorer.Run()
}
