package imports

import (
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/imports"
)

func loadRoutesTemplateImports(ctx *Context) []*imports.Import {
	ipt := map[string]*imports.Import{
		packages["fasthttp"].Name: packages["fasthttp"],
	}

	for _, m := range ctx.Methods {
		if m.HasRequiredBody {
			ipt[packages["errors"].Name] = packages["errors"]
			ipt[packages["json"].Name] = packages["json"]
		}

		if m.HasQueryArguments || m.HasHeaderArguments {
			ipt[packages["fmt"].Name] = packages["fmt"]
		}
	}

	return toSlice(ipt)
}
