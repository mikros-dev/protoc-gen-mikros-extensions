package main

import (
	"embed"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/examples/addons/domain_improve/proto"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/addon"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/addon/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	context2 "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/context"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

func loadDomainImprove(msg *context2.Message) *proto.DomainImprove {
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
	domainMessages []*context2.Message
}

func (c *Context) HasImproveDomainCall(msg *context2.Message) bool {
	if d := loadDomainImprove(msg); d != nil {
		return d.GetNewApi()
	}

	return false
}

type DomainImproveAddon struct{}

func (d *DomainImproveAddon) GetContext(ctx interface{}) interface{} {
	c := ctx.(*context2.Context)
	addonCtx := &Context{
		domainMessages: c.DomainMessages(),
	}

	return addonCtx
}

func (d *DomainImproveAddon) GetTemplateImports(_ spec.Name, _ interface{}, _ *settings.Settings) []*addon.Import {
	// Does not have imports
	return nil
}

func (d *DomainImproveAddon) GetTemplateValidator(name spec.Name, ctx interface{}) (spec.ExecutionFunc, bool) {
	c := ctx.(*context2.Context)
	pc := c.AddonContext(addonName).(*Context)

	validators := map[spec.Name]spec.ExecutionFunc{
		spec.NewName(addonName, "domain_improve"): func() bool {
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

func (d *DomainImproveAddon) Kind() spec.Kind {
	return spec.KindAPI
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
