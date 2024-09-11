package context

import (
	"google.golang.org/protobuf/compiler/protogen"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/converters"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/settings"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/template"
)

type Context struct {
	PluginName string
	ModuleName string
	Enums      []*Enum
	Methods    []*Method

	messages []*Message
	imports  map[template.Name][]*Import
	pkg      *protobuf.Protobuf
}

type BuildContextOptions struct {
	PluginName string
	Settings   *settings.Settings
	Plugin     *protogen.Plugin
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
		imports:    make(map[template.Name][]*Import),
		pkg:        pkg,
	}

	ctx.imports = LoadTemplateImports(ctx, opt.Settings)

	return ctx, nil
}

func (c *Context) GetTemplateImports(name string) []*Import {
	return c.imports[template.Name(name)]
}

func (c *Context) HasImportFor(name string) bool {
	d, ok := c.imports[template.Name(name)]
	return ok && len(d) > 0
}

func (c *Context) IsHTTPService() bool {
	return c.pkg.Service != nil && c.pkg.Service.IsHTTP()
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

// ValidateForExecution sets rules to execute or not templates while running.
func (c *Context) ValidateForExecution(name template.Name) (template.Validator, bool) {
	validators := map[template.Name]template.Validator{
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
	if c.pkg.Service != nil {
		return c.pkg.Service.Name
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
