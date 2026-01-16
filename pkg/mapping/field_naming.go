package mapping

import (
	"github.com/go-playground/validator/v10"
	"github.com/stoewer/go-strcase"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
)

// FieldNamingOptions represents the options for field naming.
type FieldNamingOptions struct {
	*FieldMappingContextOptions
}

// FieldNaming represents the naming logic for a field.
type FieldNaming struct {
	goName       string
	domainName   string
	outboundName string
	inboundName  string
}

// NewFieldNaming returns a new FieldNaming instance.
func NewFieldNaming(options *FieldNamingOptions) (*FieldNaming, error) {
	validate := options.Validate
	if validate == nil {
		validate = validator.New()
	}
	if err := validate.Struct(options); err != nil {
		return nil, err
	}

	var (
		goName = options.ProtoField.Schema.GoName
		fieldExtensions = loadFieldExtensions(options.ProtoField)
	)

	return &FieldNaming{
		goName:       goName,
		domainName:   buildDomainName(goName, fieldExtensions),
		outboundName: buildOutboundName(goName),
		inboundName:  buildInboundName(goName, fieldExtensions, loadMessageExtensions(options.ProtoMessage)),
	}, nil
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

// GoName returns the Go name of the field.
func (f *FieldNaming) GoName() string {
	return f.goName
}

// Domain returns the domain name of the field.
func (f *FieldNaming) Domain() string {
	return f.domainName
}

// Outbound returns the outbound name of the field.
func (f *FieldNaming) Outbound() string {
	return f.outboundName
}

// Inbound returns the inbound name of the field.
func (f *FieldNaming) Inbound() string {
	return f.inboundName
}
