package context

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/addon"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/mapping"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// Context represents the main context used inside the plugin template files.
type Context struct {
	PluginName string
	ModuleName string
	Enums      []*Enum
	Methods    []*Method
	Package    *protobuf.Protobuf

	messages []*Message
	imports  map[spec.Name][]*templateImport
	addons   map[string]*addon.Addon
	settings *settings.Settings
}

// BuildContextOptions represents the options used to build the context.
type BuildContextOptions struct {
	PluginName string
	Settings   *settings.Settings
	Plugin     *protogen.Plugin
	Addons     []*addon.Addon
}

// BuildContext builds the context from the protobuf file(s).
func BuildContext(opt BuildContextOptions) (*Context, error) {
	// Handle the protobuf file(s)
	pkg, err := protobuf.Parse(protobuf.ParseOptions{
		Plugin: opt.Plugin,
	})
	if err != nil {
		return nil, err
	}

	// And build the templates context
	messages, err := loadMessages(pkg, loadMessagesOptions{
		Settings: opt.Settings,
	})
	if err != nil {
		return nil, err
	}

	methods, err := loadMethods(pkg, messages, opt.Settings)
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

// GetTemplateImports returns the imports for the given template name.
func (c *Context) GetTemplateImports(name string) []*templateImport {
	return c.imports[spec.Name(name)]
}

// GetAddonTemplateImports returns the imports for the given addon template name.
func (c *Context) GetAddonTemplateImports(addonName, tplName string) []*templateImport {
	if a, ok := c.addons[addonName]; ok {
		var (
			ipt          = a.Addon().GetTemplateImports(spec.Name(tplName), c, c.settings)
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

// HasImportFor returns true if the given template has an import for the given
// name.
func (c *Context) HasImportFor(name string) bool {
	d, ok := c.imports[spec.Name(name)]
	return ok && len(d) > 0
}

// HasAddonImportFor returns true if the given addon has an import for the given
// template name.
func (c *Context) HasAddonImportFor(addonName, tplName string) bool {
	if a, ok := c.addons[addonName]; ok {
		return len(a.Addon().GetTemplateImports(spec.Name(tplName), c, c.settings)) > 0
	}

	return false
}

// IsHTTPService returns true if the current package is an HTTP service.
func (c *Context) IsHTTPService() bool {
	return c.Package.Service != nil && c.Package.Service.IsHTTP()
}

// DomainMessages returns the messages that should be exported as domain.
func (c *Context) DomainMessages() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		// Every wire message will have Domain equivalents
		if m.Type == mapping.Wire && m.DomainExport() {
			messages = append(messages, m)
		}
	}

	return messages
}

// WireInputMessages returns the messages that should be exported as wire input.
func (c *Context) WireInputMessages() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		isWireInput := m.Type == mapping.WireInput || manualExportToWireInput(m)
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

// OutboundMessages returns the messages that should be exported as outbound.
func (c *Context) OutboundMessages() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		if m.OutboundExport() {
			messages = append(messages, m)
		}
	}

	return messages
}

// CustomAPIExtensions returns the messages that have custom API defined in them.
func (c *Context) CustomAPIExtensions() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		if m.HasCustomAPICodeExtension() {
			messages = append(messages, m)
		}
	}

	return messages
}

// GetTemplateValidator returns the validator for the given template name.
func (c *Context) GetTemplateValidator(name spec.Name, _ interface{}) (spec.ExecutionFunc, bool) {
	validators := map[spec.Name]spec.ExecutionFunc{
		spec.NewName("api", "domain"): func() bool {
			return len(c.DomainMessages()) > 0
		},
		spec.NewName("api", "enum"): func() bool {
			return len(c.Enums) > 0
		},
		spec.NewName("api", "custom_api"): func() bool {
			return len(c.CustomAPIExtensions()) > 0
		},
		spec.NewName("api", "http_server"): func() bool {
			return c.IsHTTPService()
		},
		spec.NewName("api", "routes"): func() bool {
			return c.IsHTTPService()
		},
		spec.NewName("api", "outbound"): func() bool {
			return c.IsHTTPService() || len(c.OutboundMessages()) > 0
		},
		spec.NewName("api", "wire"): func() bool {
			return len(c.DomainMessages()) > 0
		},
		spec.NewName("api", "wire_input"): func() bool {
			return len(c.WireInputMessages()) > 0
		},
		spec.NewName("api", "common"): func() bool {
			return c.UseCommonConverters() || c.OutboundHasBitflagField()
		},
		spec.NewName("api", "validation"): func() bool {
			return c.HasValidatableMessage()
		},
		spec.NewName("testing", "testing"): func() bool {
			return len(c.DomainMessages()) > 0 && c.settings.Templates.Test
		},
		spec.NewName("testing", "http_server"): func() bool {
			return c.IsHTTPService() && c.settings.Templates.Test
		},
	}

	v, ok := validators[name]
	return v, ok
}

// Extension returns the extension for the generated source files.
func (c *Context) Extension() string {
	return "go"
}

// ServiceName returns the name of the service associated with the context.
func (c *Context) ServiceName() string {
	if c.Package.Service != nil {
		return c.Package.Service.Name
	}

	return c.ModuleName
}

// HasRequiredBody returns true if the service has any method with a required
// body.
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

// OutboundHasBitflagField returns true if the service has any outbound message
// with a bitflag field.
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

// HasValidatableMessage returns true if the service has any message with a
// validatable field.
func (c *Context) HasValidatableMessage() bool {
	return len(c.ValidatableMessages()) > 0
}

// ValidatableMessages returns the messages that have a validatable field.
func (c *Context) ValidatableMessages() []*Message {
	var messages []*Message
	for _, m := range c.messages {
		if m.HasValidatableField() || m.Type == mapping.WireInput {
			messages = append(messages, m)
		}
	}

	return messages
}

// AddonContext returns the context for the given addon.
func (c *Context) AddonContext(addonName string) interface{} {
	if a, ok := c.addons[addonName]; ok {
		return a.Addon().GetContext(c)
	}

	return nil
}

// UseCommonConverters returns true if the common converters defined inside the
// settings should be used.
func (c *Context) UseCommonConverters() bool {
	if c.settings.Templates.Common != nil {
		return c.settings.Templates.Common.Converters
	}

	return false
}

// HasAddonIntoOutboundExtensionContent returns true if the given message has
// an addon with custom outbound extension content.
func (c *Context) HasAddonIntoOutboundExtensionContent(msg *Message) bool {
	for _, a := range c.addons {
		if ext := a.OutboundExtension(); ext != nil && ext.IntoOutbound(msg, "r") != "" {
			return true
		}
	}

	return false
}

// AddonIntoOutboundExtensionContent returns the custom outbound extension
// content for the given message.
func (c *Context) AddonIntoOutboundExtensionContent(msg *Message, receiver string) string {
	var output string
	for _, a := range c.addons {
		if ext := a.OutboundExtension(); ext != nil {
			output += ext.IntoOutbound(msg, receiver)
		}
	}

	return output
}
