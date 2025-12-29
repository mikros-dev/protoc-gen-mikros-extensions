package main

import (
	"embed"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/addon"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/context"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

//go:embed *.tmpl
var templateFiles embed.FS

type WireImproveAddon struct{}

func (w *WireImproveAddon) GetContext(_ interface{}) interface{} {
	return struct {
		Example string
	}{
		Example: "Hello world!",
	}
}

func (w *WireImproveAddon) GetTemplateImports(name spec.Name, _ interface{}, _ *settings.Settings) []*addon.Import {
	ipt := map[spec.Name][]*addon.Import{
		spec.NewName(addonName, "wire_improve"): {
			{
				Name: "fmt",
			},
		},
	}

	if i, ok := ipt[name]; ok {
		return i
	}

	return nil
}

func (w *WireImproveAddon) GetTemplateValidator(name spec.Name, ctx interface{}) (spec.ExecutionFunc, bool) {
	c := ctx.(*context.Context)

	validators := map[spec.Name]spec.ExecutionFunc{
		spec.NewName(addonName, "wire_improve"): func() bool {
			return len(c.DomainMessages()) > 0
		},
	}

	v, ok := validators[name]
	return v, ok
}

func (w *WireImproveAddon) Kind() spec.Kind {
	return spec.KindAPI
}

func (w *WireImproveAddon) Templates() embed.FS {
	return templateFiles
}

func (w *WireImproveAddon) Name() string {
	return addonName
}

var (
	// Addon is the addon exported type that implements the supported interface
	// to be a valid plugin addon.
	Addon     WireImproveAddon
	addonName = "custom-wire-api"
)
