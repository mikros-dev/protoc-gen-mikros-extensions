package imports

import (
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/settings"
)

func loadDomainTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	ipt := make(map[string]*Import)

	for k, v := range loadImportsFromMessages(ctx, cfg, ctx.DomainMessages) {
		ipt[k] = v
	}

	return toSlice(ipt)
}
