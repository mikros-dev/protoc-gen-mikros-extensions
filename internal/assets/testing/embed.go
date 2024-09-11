package testing

import (
	"embed"
)

//go:embed *.tmpl
var Files embed.FS
