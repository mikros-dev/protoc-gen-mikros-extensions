package testing

import (
	"embed"
)

// Files gathers all templates files for testing package generated for protobuf
// modules.
//
//go:embed *.tmpl
var Files embed.FS
