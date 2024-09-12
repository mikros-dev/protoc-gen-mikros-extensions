package imports

import (
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/settings"
)

func loadWireInputTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	for k, v := range loadImportsFromMessages(ctx, cfg, ctx.WireInputMessages) {
		imports[k] = v
	}

	return toSlice(imports)
}
