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
	goName       string
	domainName   string
	outboundName string
	inboundName  string
}

func newFieldNaming(options *FieldNameOptions) *FieldNaming {
	return &FieldNaming{
		goName:       options.GoName,
		domainName:   buildDomainName(options.GoName, options.FieldExtensions),
		outboundName: buildOutboundName(options.GoName),
		inboundName:  buildInboundName(options.GoName, options.FieldExtensions, options.MessageExtensions),
	}
}

func buildDomainName(goName string, ext *extensions.MikrosFieldExtensions) string {
	if domain := ext.GetDomain(); domain != nil {
		if n := domain.GetName(); n != "" {
			return strcase.UpperCamelCase(n)
		}
	}

	return goName
}

func buildOutboundName(goName string) string {
	return goName
}

func buildInboundName(
	goName string,
	ext *extensions.MikrosFieldExtensions,
	msgExt *extensions.MikrosMessageExtensions,
) string {
	name := goName
	if inbound := ext.GetInbound(); inbound != nil {
		if n := inbound.GetName(); n != "" {
			name = n
		}
	}

	// The default is snake_case
	fieldName := strcase.SnakeCase(name)
	if messageInbound := msgExt.GetInbound(); messageInbound != nil {
		if messageInbound.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
			fieldName = inboundOutboundCamelCase(name)
		}
	}

	return fieldName
}

func inboundOutboundCamelCase(s string) string {
	return strcase.LowerCamelCase(s)
}

func (f *FieldNaming) GoName() string {
	return f.goName
}

func (f *FieldNaming) Domain() string {
	return f.domainName
}

func (f *FieldNaming) Outbound() string {
	return f.outboundName
}

func (f *FieldNaming) Inbound() string {
	return f.inboundName
}
