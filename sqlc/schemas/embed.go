package schemas

import (
	"embed"
)

//go:embed *.sql
var FS embed.FS
