//+build fabric_sample

package ui

import (
	"embed"
)

const Prefix = "dist_fabric_sample"

//go:embed dist_fabric_sample
var FS embed.FS
