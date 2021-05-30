package fabric_sample

import (
	"explorer"
	"explorer/hf"
	"explorer/pg"
)

func init() {
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
	hf.RegisterJSONValueType("basic", "^asset", "asset")
}
