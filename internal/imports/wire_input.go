package imports

import (
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/imports"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/settings"
)

func loadWireInputTemplateImports(ctx *Context, cfg *settings.Settings) []*imports.Import {
	ipt := make(map[string]*imports.Import)

	for k, v := range loadImportsFromMessages(ctx, cfg, ctx.WireInputMessages) {
		ipt[k] = v
	}

	return toSlice(ipt)
}
