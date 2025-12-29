package api

import (
	"embed"
)

// Files gathers all templates files for API improvements of a protobuf file.
//
//go:embed *.tmpl
var Files embed.FS
