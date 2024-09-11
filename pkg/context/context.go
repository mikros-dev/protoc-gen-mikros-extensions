package context

import (
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/settings"
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/converters"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/imports"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/addon"
	mimports "github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/imports"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/template"
)

type Context struct {
	PluginName string
	ModuleName string
	Enums      []*Enum
	Methods    []*Method
	Package    *protobuf.Protobuf

	messages []*Message
	imports  map[template.Name][]*mimports.Import
	addons   map[string]addon.Addon
	settings *settings.Settings
}

type BuildContextOptions struct {
	PluginName string
	Settings   *settings.Settings
	Plugin     *protogen.Plugin
	Addons     []addon.Addon
}

func BuildContext(opt BuildContextOptions) (*Context, error) {
	// Handle the protobuf file(s)
	pkg, err := protobuf.Parse(protobuf.ParseOptions{
		Plugin: opt.Plugin,
	})
	if err != nil {
		return nil, err
	}

	// And build the templates context
	messages, err := loadMessages(pkg, LoadMessagesOptions{
		Settings: opt.Settings,
	})
	if err != nil {
		return nil, err
	}

	methods, err := loadMethods(pkg, messages)
	if err != nil {
		return nil, err
	}

	ctx := &Context{
		PluginName: opt.PluginName,
		ModuleName: pkg.ModuleName,
		Enums:      loadEnums(pkg),
		Methods:    methods,
		messages:   messages,
		Package:    pkg,
		settings:   opt.Settings,
	}

	addons := make(map[string]addon.Addon)
	for _, a := range opt.Addons {
		addons[a.Name()] = a
	}

	ctx.addons = addons
	ctx.imports = imports.LoadTemplateImports(toImportsContext(ctx), opt.Settings)

	return ctx, nil
}

func (c *Context) GetTemplateImports(name string) []*mimports.Import {
	return c.imports[template.Name(name)]
}

func (c *Context) GetAddonTemplateImports(addonName, tplName string) []*mimports.Import {
	if a, ok := c.addons[addonName]; ok {
		return a.GetTemplateImports(template.Name(tplName), c, c.settings)
	}

	return nil
}

func (c *Context) HasImportFor(name string) bool {
	d, ok := c.imports[template.Name(name)]
	return ok && len(d) > 0
}

func (c *Context) HasAddonImportFor(addonName, tplName string) bool {
	if a, ok := c.addons[addonName]; ok {
		return len(a.GetTemplateImports(template.Name(tplName), c, c.settings)) > 0
	}

	return false
}

func (c *Context) IsHTTPService() bool {
	return c.Package.Service != nil && c.Package.Service.IsHTTP()
}

func (c *Context) DomainMessages() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		// Every wire message will have Domain equivalents
		if m.Type == converters.WireMessage && m.DomainExport() {
			messages = append(messages, m)
		}
	}

	return messages
}

func (c *Context) WireInputMessages() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		if m.Type == converters.WireInputMessage && m.DomainExport() {
			messages = append(messages, m)
		}
	}

	return messages
}

func (c *Context) OutboundMessages() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		if m.OutboundExport() {
			messages = append(messages, m)
		}
	}

	return messages
}

func (c *Context) WireExtensions() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		if m.HasWireCustomCodeExtension() {
			messages = append(messages, m)
		}
	}

	return messages
}

func (c *Context) GetTemplateValidator(name template.Name, _ interface{}) (template.ValidateForExecution, bool) {
	validators := map[template.Name]template.ValidateForExecution{
		template.NewName("api", "domain"): func() bool {
			return len(c.DomainMessages()) > 0
		},
		template.NewName("api", "enum"): func() bool {
			return len(c.Enums) > 0
		},
		template.NewName("api", "wire"): func() bool {
			return len(c.WireExtensions()) > 0
		},
		template.NewName("api", "http_server"): func() bool {
			return c.IsHTTPService()
		},
		template.NewName("api", "routes"): func() bool {
			return c.IsHTTPService()
		},
		template.NewName("api", "outbound"): func() bool {
			return c.IsHTTPService() || len(c.OutboundMessages()) > 0
		},
		template.NewName("api", "wire_input"): func() bool {
			return len(c.WireInputMessages()) > 0 && c.IsHTTPService()
		},
		template.NewName("api", "common"): func() bool {
			return c.HasProtobufValueField() || c.OutboundHasBitflagField()
		},
		template.NewName("api", "validation"): func() bool {
			return c.HasValidatableMessage()
		},
		template.NewName("testing", "testing"): func() bool {
			return len(c.DomainMessages()) > 0
		},
		template.NewName("testing", "http_server"): func() bool {
			return c.IsHTTPService()
		},
	}

	v, ok := validators[name]
	return v, ok
}

func (c *Context) Extension() string {
	return "go"
}

func (c *Context) ServiceName() string {
	if c.Package.Service != nil {
		return c.Package.Service.Name
	}

	return c.ModuleName
}

func (c *Context) HasRequiredBody() bool {
	if len(c.Methods) > 0 {
		for _, m := range c.Methods {
			if m.HasRequiredBody() {
				return true
			}
		}
	}

	return false
}

func (c *Context) OutboundHasBitflagField() bool {
	if len(c.OutboundMessages()) > 0 {
		for _, m := range c.OutboundMessages() {
			if m.HasBitflagField() {
				return true
			}
		}
	}

	return false
}

func (c *Context) HasProtobufValueField() bool {
	for _, m := range c.messages {
		if m.HasProtobufValueField() {
			return true
		}
	}

	return false
}

func (c *Context) HasValidatableMessage() bool {
	return len(c.ValidatableMessages()) > 0
}

func (c *Context) ValidatableMessages() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		if m.HasValidatableField() {
			messages = append(messages, m)
		}
	}

	return messages
}
