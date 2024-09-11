package imports

import (
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/imports"
)

func loadHttpServerTemplateImports() []*imports.Import {
	ipt := map[string]*imports.Import{
		packages["context"].Name:         packages["context"],
		packages["errors"].Name:          packages["errors"],
		packages["fasthttp"].Name:        packages["fasthttp"],
		packages["fasthttp-router"].Name: packages["fasthttp-router"],
	}

	return toSlice(ipt)
}

func loadTestingHttpServerTemplateImports(ctx *Context) []*imports.Import {
	ipt := map[string]*imports.Import{
		packages["fasthttp-router"].Name: packages["fasthttp-router"],
		ctx.ModuleName:                   importAnotherModule(ctx.ModuleName, ctx.ModuleName, ctx.FullPath),
	}

	return toSlice(ipt)
}
