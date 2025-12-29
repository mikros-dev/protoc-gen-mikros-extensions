package imports

import (
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// Common represents the 'api/common.tmpl' importer.
type Common struct{}

// Name returns the template name.
func (c *Common) Name() spec.Name {
	return spec.NewName("api", "common")
}

// Load returns a slice of imports for the template.
func (c *Common) Load(ctx *Context, _ *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	if ctx.OutboundHasBitflagField {
		imports["strings"] = packages["strings"]
	}

	if ctx.UseCommonConverters {
		imports["time"] = packages["time"]
		imports["protostruct"] = packages["protostruct"]
		imports["prototimestamp"] = packages["prototimestamp"]
	}

	return toSlice(imports)
}
