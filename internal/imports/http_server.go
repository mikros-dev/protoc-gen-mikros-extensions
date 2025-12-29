package imports

import (
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// HTTPServer represents the 'api/http_server.tmpl' importer
type HTTPServer struct{}

// Name returns the template name.
func (h *HTTPServer) Name() spec.Name {
	return spec.NewName("api", "http_server")
}

// Load returns a slice of imports for the template.
func (h *HTTPServer) Load(_ *Context, _ *settings.Settings) []*Import {
	imports := map[string]*Import{
		packages["context"].Name:         packages["context"],
		packages["errors"].Name:          packages["errors"],
		packages["fasthttp"].Name:        packages["fasthttp"],
		packages["fasthttp-router"].Name: packages["fasthttp-router"],
	}

	return toSlice(imports)
}

// TestingHTTPServer represents the 'testing/http_server.tmpl' importer
type TestingHTTPServer struct{}

// Name returns the template name.
func (t *TestingHTTPServer) Name() spec.Name {
	return spec.NewName("testing", "http_server")
}

// Load returns a slice of imports for the template.
func (t *TestingHTTPServer) Load(ctx *Context, _ *settings.Settings) []*Import {
	imports := map[string]*Import{
		packages["fasthttp-router"].Name: packages["fasthttp-router"],
		ctx.ModuleName:                   importAnotherModule(ctx.ModuleName, ctx.ModuleName, ctx.FullPath),
	}

	return toSlice(imports)
}
