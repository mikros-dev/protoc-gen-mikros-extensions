package imports

import (
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// Enum represents the 'api/enum.tmpl' importer.
type Enum struct{}

// Name returns the template name.
func (e *Enum) Name() spec.Name {
	return spec.NewName("api", "enum")
}

// Load returns a slice of imports for the template.
func (e *Enum) Load(_ *Context, _ *settings.Settings) []*Import {
	imports := map[string]*Import{
		"strings": packages["strings"],
	}

	return toSlice(imports)
}
