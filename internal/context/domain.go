package context

import (
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/settings"
)

func loadDomainTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	for k, v := range loadImportsFromMessages(ctx, cfg, ctx.DomainMessages()) {
		imports[k] = v
	}

	return toSlice(imports)
}
