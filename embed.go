package explorer

import "embed"

//go:embed service.swagger.json
var FS embed.FS

const SwaggerFile = "service.swagger.json"
