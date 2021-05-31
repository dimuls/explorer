package main

import (
	"explorer"
	"explorer/example/fabric-sample/ui"
	"explorer/hf"
	"explorer/pg"
)

func main() {
	hf.RegisterJSONValueType("basic", "^asset", "asset")

	pg.RegisterUI(ui.FS)

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
