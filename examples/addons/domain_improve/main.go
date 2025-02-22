package main

import (
	"embed"

	"google.golang.org/protobuf/proto"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/mikros/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/addon"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/context"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	tpl_types "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/types"
)

func loadDomainImprove(msg *descriptor.DescriptorProto) *extensions.DomainImprove {
	if msg.Options != nil {
		v := proto.GetExtension(msg.Options, extensions.E_Improve)
		if val, ok := v.(*extensions.DomainImprove); ok {
			return val
		}
	}

	return nil
}

//go:embed *.tmpl
var templateFiles embed.FS

type Context struct {
	domainMessages []*context.Message
}

func (c *Context) HasImproveDomainCall(msg *context.Message) bool {
	if d := loadDomainImprove(msg.ProtoMessage.Proto); d != nil {
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
		tpl_types.NewName(addonName, "domain_improve"): func() bool {
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
	return tpl_types.KindApi
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
