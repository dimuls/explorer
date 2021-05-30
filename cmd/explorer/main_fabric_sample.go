//+build fabric_sample

package main

import (
	"explorer"
	"explorer/hf"
	"explorer/pg"
)

func run() {

	hf.RegisterJSONValueType("basic", "^asset", "asset")

	pg.RegisterQuery(pg.Query{
		Name: "assets_by_owner_and_size",
		Body: `
			select *
			from state
			where type = 'asset'
				and value#>>'{"owner"}' = $1
				and (value#>>'{"size"}')::integer < $2
			order by key asc
		`,
		DestElem: &explorer.State{},
	})

	pg.RunExplorer()
}
