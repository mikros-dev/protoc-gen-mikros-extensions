package imports

import (
	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
)

func loadWireTemplateImports(ctx *Context) []*Import {
	imports := make(map[string]*Import)

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

				imports[key] = &Import{
					Alias: i.GetAlias(),
					Name:  i.GetName(),
				}
			}
		}
	}

	return toSlice(imports)
}
