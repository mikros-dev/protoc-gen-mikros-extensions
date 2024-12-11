package main

import (
	"embed"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/addon"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/context"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template"
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

func (w *WireImproveAddon) GetTemplateImports(name template.Name, _ interface{}, _ *settings.Settings) []*addon.Import {
	ipt := map[template.Name][]*addon.Import{
		template.NewName(template.KindGo, "wire_improve"): {
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

func (w *WireImproveAddon) GetTemplateValidator(name template.Name, ctx interface{}) (template.ValidateForExecution, bool) {
	c := ctx.(*context.Context)

	validators := map[template.Name]template.ValidateForExecution{
		template.NewName(template.KindGo, "wire_improve"): func() bool {
			return len(c.DomainMessages()) > 0
		},
	}

	v, ok := validators[name]
	return v, ok
}

func (w *WireImproveAddon) Kind() template.Kind {
	return template.KindGo
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
