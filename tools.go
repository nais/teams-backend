//go:build tools
// +build tools

package tools

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/kyleconroy/sqlc/cmd/sqlc"
	_ "github.com/vektra/mockery/v2"
	_ "mvdan.cc/gofumpt"
)
