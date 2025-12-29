package imports

import (
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// Routes represents the 'api/routes.tmpl' importer
type Routes struct{}

// Name returns the template name.
func (r *Routes) Name() spec.Name {
	return spec.NewName("api", "routes")
}

// Load returns a slice of imports for the template.
func (r *Routes) Load(ctx *Context, _ *settings.Settings) []*Import {
	imports := map[string]*Import{
		packages["fasthttp"].Name: packages["fasthttp"],
		packages["fmt"].Name:      packages["fmt"],
	}

	for _, m := range ctx.Methods {
		if m.HasRequiredBody {
			imports[packages["errors"].Name] = packages["errors"]
			imports[packages["json"].Name] = packages["json"]
		}

		if m.HasQueryArguments || m.HasHeaderArguments {
			imports[packages["fmt"].Name] = packages["fmt"]
		}
	}

	return toSlice(imports)
}
