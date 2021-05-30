//+build fabric_sample

package main

import (
	_ "explorer/mod/fabric_sample"
	"explorer/pg"
)

func run() {
	pg.RunExplorer()
}
