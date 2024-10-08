package imports

import (
	"github.com/mikros-dev/protoc-gen-mikros-extensions/mikros/extensions"
)

func loadWireTemplateImports(ctx *Context) []*Import {
	imports := make(map[string]*Import)

	for _, m := range ctx.WireExtensions {
		ext := extensions.LoadMessageExtensions(m.ProtoMessage.Proto)
		if ext == nil {
			continue
		}

		options := ext.GetWire()
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
