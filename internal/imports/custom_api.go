package imports

import (
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// CustomAPI represents the 'api/custom_api.tmpl' importer.
type CustomAPI struct{}

// Name returns the template name.
func (c *CustomAPI) Name() spec.Name {
	return spec.NewName("api", "custom_api")
}

// Load returns a slice of imports for the template.
func (c *CustomAPI) Load(ctx *Context, _ *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	for _, m := range ctx.WireExtensions {
		ext := extensions.LoadMessageExtensions(m.ProtoMessage.Proto)
		if ext == nil {
			continue
		}

		options := ext.GetCustomApi()
		if options == nil {
			continue
		}

		for _, c := range options.GetFunction() {
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
