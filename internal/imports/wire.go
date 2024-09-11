package imports

import (
	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/imports"
)

func loadWireTemplateImports(ctx *Context) []*imports.Import {
	ipt := make(map[string]*imports.Import)

	for _, m := range ctx.WireExtensions {
		options := extensions.LoadMessageWireExtensionOptions(m.ProtoMessage.Proto)
		if options == nil {
			continue
		}

		for _, c := range options.GetCustomCode() {
			for _, i := range c.GetImport() {
				key := i.GetName()
				if key == "" {
					key = i.GetName()
				}

				ipt[key] = &imports.Import{
					Alias: i.GetAlias(),
					Name:  i.GetName(),
				}
			}
		}
	}

	return toSlice(ipt)
}
