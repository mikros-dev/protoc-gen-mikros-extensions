package main

import (
	"github.com/bufbuild/protoplugin"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/plugin"
)

func main() {
	protoplugin.Main(
		protoplugin.HandlerFunc(plugin.Handle),
	)
}
