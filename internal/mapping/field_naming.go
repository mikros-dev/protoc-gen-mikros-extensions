package mapping

import (
	"github.com/stoewer/go-strcase"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
)

type FieldNameOptions struct {
	GoName            string
	FieldExtensions   *extensions.MikrosFieldExtensions
	MessageExtensions *extensions.MikrosMessageExtensions
}

type FieldNaming struct {
	GoName            string
	extensions        *extensions.MikrosFieldExtensions
	messageExtensions *extensions.MikrosMessageExtensions
}

func newFieldNaming(options *FieldNameOptions) *FieldNaming {
	return &FieldNaming{
		GoName:            options.GoName,
		extensions:        options.FieldExtensions,
		messageExtensions: options.MessageExtensions,
	}
}

func (f *FieldNaming) DomainName() string {
	if f.extensions != nil {
		if domain := f.extensions.GetDomain(); domain != nil {
			if n := domain.GetName(); n != "" {
				return strcase.UpperCamelCase(n)
			}
		}
	}

	return f.GoName
}

func (f *FieldNaming) ResolveDomainNameForTag(fieldName string) string {
	// The default is snake_case
	fieldName = strcase.SnakeCase(fieldName)

	if f.messageExtensions != nil {
		if messageDomain := f.messageExtensions.GetDomain(); messageDomain != nil {
			if messageDomain.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
				fieldName = strcase.LowerCamelCase(fieldName)
			}
		}
	}

	return fieldName
}

func (f *FieldNaming) OutboundName() string {
	return f.GoName
}

func (f *FieldNaming) ResolveOutboundNameForTag(fieldName string) string {
	fieldName = strcase.SnakeCase(fieldName)

	if f.messageExtensions != nil {
		if messageOutbound := f.messageExtensions.GetOutbound(); messageOutbound != nil {
			if messageOutbound.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
				fieldName = inboundOutboundCamelCase(fieldName)
			}
		}
	}

	return fieldName
}

func (f *FieldNaming) OutboundJSONTagFieldName() string {
	name := f.DomainName()
	if f.extensions != nil {
		if outbound := f.extensions.GetOutbound(); outbound != nil {
			if n := outbound.GetName(); n != "" {
				name = n
			}
		}
	}

	// The default is snake_case
	fieldName := strcase.SnakeCase(name)
	if f.messageExtensions != nil {
		if messageOutbound := f.messageExtensions.GetOutbound(); messageOutbound != nil {
			if messageOutbound.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
				fieldName = inboundOutboundCamelCase(name)
			}
		}
	}

	return fieldName
}

func (f *FieldNaming) InboundName() string {
	name := f.DomainName()
	if f.extensions != nil {
		if inbound := f.extensions.GetInbound(); inbound != nil {
			if n := inbound.GetName(); n != "" {
				name = n
			}
		}
	}

	// The default is snake_case
	fieldName := strcase.SnakeCase(name)
	if f.messageExtensions != nil {
		if messageInbound := f.messageExtensions.GetInbound(); messageInbound != nil {
			if messageInbound.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
				fieldName = inboundOutboundCamelCase(name)
			}
		}
	}

	return fieldName
}

func inboundOutboundCamelCase(s string) string {
	return strcase.LowerCamelCase(s)
}
