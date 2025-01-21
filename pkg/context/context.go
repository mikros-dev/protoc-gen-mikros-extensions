package context

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/addon"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/mikros/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/converters"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template"
)

type Context struct {
	PluginName   string
	ModuleName   string
	TemplateKind template.Kind
	Enums        []*Enum
	Methods      []*Method
	Package      *protobuf.Protobuf

	messages []*Message
	imports  map[template.Name][]*templateImport
	addons   map[string]*addon.Addon
	settings *settings.Settings
}

type BuildContextOptions struct {
	PluginName string
	Settings   *settings.Settings
	Plugin     *protogen.Plugin
	Addons     []*addon.Addon
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

	addons := make(map[string]*addon.Addon)
	for _, a := range opt.Addons {
		addons[a.Addon().Name()] = a
	}

	ctx.addons = addons
	ctx.imports = loadImports(ctx, opt.Settings)

	return ctx, nil
}

func (c *Context) GetTemplateImports(name string) []*templateImport {
	return c.imports[template.Name(name)]
}

func (c *Context) GetAddonTemplateImports(addonName, tplName string) []*templateImport {
	if a, ok := c.addons[addonName]; ok {
		var (
			ipt          = a.Addon().GetTemplateImports(template.Name(tplName), c, c.settings)
			addonImports = make([]*templateImport, len(ipt))
		)

		for i, ii := range ipt {
			addonImports[i] = &templateImport{
				Alias: ii.Alias,
				Name:  ii.Name,
			}
		}

		return addonImports
	}

	return nil
}

func (c *Context) HasImportFor(name string) bool {
	d, ok := c.imports[template.Name(name)]
	return ok && len(d) > 0
}

func (c *Context) HasAddonImportFor(addonName, tplName string) bool {
	if a, ok := c.addons[addonName]; ok {
		return len(a.Addon().GetTemplateImports(template.Name(tplName), c, c.settings)) > 0
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
		isWireInput := m.Type == converters.WireInputMessage || manualExportToWireInput(m)
		if isWireInput && m.DomainExport() {
			messages = append(messages, m)
		}
	}

	return messages
}

func manualExportToWireInput(m *Message) bool {
	if ext := extensions.LoadMessageExtensions(m.ProtoMessage.Proto); ext != nil {
		if wireInput := ext.GetWireInput(); wireInput != nil {
			return wireInput.GetExport()
		}
	}

	return false
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

func (c *Context) CustomApiExtensions() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		if m.HasCustomApiCodeExtension() {
			messages = append(messages, m)
		}
	}

	return messages
}

func (c *Context) SetTemplateKind(kind template.Kind) {
	c.TemplateKind = kind
}

func (c *Context) GetTemplateValidator(name template.Name, _ interface{}) (template.ValidateForExecution, bool) {
	var (
		golang  = template.KindGo
		testing = template.KindTest
		rust    = template.KindRust
	)

	validators := map[template.Name]template.ValidateForExecution{
		template.NewName(golang, "domain"): func() bool {
			return len(c.DomainMessages()) > 0
		},
		template.NewName(golang, "enum"): func() bool {
			return len(c.Enums) > 0
		},
		template.NewName(golang, "custom_api"): func() bool {
			return len(c.CustomApiExtensions()) > 0
		},
		template.NewName(golang, "http_server"): func() bool {
			return c.IsHTTPService()
		},
		template.NewName(golang, "routes"): func() bool {
			return c.IsHTTPService()
		},
		template.NewName(golang, "outbound"): func() bool {
			return c.IsHTTPService() || len(c.OutboundMessages()) > 0
		},
		template.NewName(golang, "wire_input"): func() bool {
			return len(c.WireInputMessages()) > 0
		},
		template.NewName(golang, "common"): func() bool {
			return c.UseCommonConverters() || c.OutboundHasBitflagField()
		},
		template.NewName(golang, "validation"): func() bool {
			return c.HasValidatableMessage()
		},
		template.NewName(testing, "testing"): func() bool {
			return len(c.DomainMessages()) > 0 && c.settings.Templates.Test.IsEnabled()
		},
		template.NewName(testing, "http_server"): func() bool {
			return c.IsHTTPService() && c.settings.Templates.Test.IsEnabled()
		},
		template.NewName(rust, "router.rs"): func() bool {
			return c.IsHTTPService()
		},
		template.NewName(rust, "mod.rs"): func() bool {
			return c.IsHTTPService()
		},
	}

	v, ok := validators[name]
	return v, ok
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

func (c *Context) HasValidatableMessage() bool {
	return len(c.ValidatableMessages()) > 0
}

func (c *Context) ValidatableMessages() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		if m.HasValidatableField() || m.Type == converters.WireInputMessage {
			messages = append(messages, m)
		}
	}

	return messages
}

func (c *Context) AddonContext(addonName string) interface{} {
	if a, ok := c.addons[addonName]; ok {
		return a.Addon().GetContext(c)
	}

	return nil
}

func (c *Context) UseCommonConverters() bool {
	if c.settings.Templates.Common != nil {
		return c.settings.Templates.Common.Converters
	}

	return false
}

func (c *Context) HasAddonIntoOutboundExtensionContent(msg *Message) bool {
	for _, a := range c.addons {
		if ext := a.OutboundExtension(); ext != nil && ext.IntoOutbound(msg, "r") != "" {
			return true
		}
	}

	return false
}

func (c *Context) AddonIntoOutboundExtensionContent(msg *Message, receiver string) string {
	var output string
	for _, a := range c.addons {
		if ext := a.OutboundExtension(); ext != nil {
			output += ext.IntoOutbound(msg, receiver)
		}
	}

	return output
}
