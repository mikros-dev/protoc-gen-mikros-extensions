package main

import (
	"embed"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/addon"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/addon/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/context"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	tpl_types "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/types"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/examples/addons/domain_improve/proto"
)

func loadDomainImprove(msg *context.Message) *proto.DomainImprove {
	if extensions.HasExtension(msg.ProtoMessage.Proto.GetOptions(), proto.E_Improve.TypeDescriptor()) {
		var domainImprove proto.DomainImprove
		if err := extensions.RetrieveExtension(msg.ProtoMessage.Proto.GetOptions(), proto.E_Improve.TypeDescriptor(), &domainImprove); err != nil {
			return nil
		}

		return &domainImprove
	}

	return nil
}

//go:embed *.tmpl
var templateFiles embed.FS

type Context struct {
	domainMessages []*context.Message
}

func (c *Context) HasImproveDomainCall(msg *context.Message) bool {
	if d := loadDomainImprove(msg); d != nil {
		return d.GetNewApi()
	}

	return false
}

type DomainImproveAddon struct{}

func (d *DomainImproveAddon) GetContext(ctx interface{}) interface{} {
	c := ctx.(*context.Context)
	addonCtx := &Context{
		domainMessages: c.DomainMessages(),
	}

	return addonCtx
}

func (d *DomainImproveAddon) GetTemplateImports(_ tpl_types.Name, _ interface{}, _ *settings.Settings) []*addon.Import {
	// Does not have imports
	return nil
}

func (d *DomainImproveAddon) GetTemplateValidator(name tpl_types.Name, ctx interface{}) (tpl_types.ValidateForExecution, bool) {
	c := ctx.(*context.Context)
	pc := c.AddonContext(addonName).(*Context)

	validators := map[tpl_types.Name]tpl_types.ValidateForExecution{
		tpl_types.NewName(tpl_types.KindGo, "domain_improve"): func() bool {
			for _, msg := range c.DomainMessages() {
				if pc.HasImproveDomainCall(msg) {
					return true
				}
			}

			return false
		},
	}

	v, ok := validators[name]
	return v, ok
}

func (d *DomainImproveAddon) Kind() tpl_types.Kind {
	return tpl_types.KindGo
}

func (d *DomainImproveAddon) Templates() embed.FS {
	return templateFiles
}

func (d *DomainImproveAddon) Name() string {
	return addonName
}

var (
	// Addon is the addon exported type that implements the supported interface
	// to be a valid plugin addon.
	Addon     DomainImproveAddon
	addonName = "custom-domain-api"
)

func main() {}
