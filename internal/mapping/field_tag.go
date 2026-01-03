package mapping

import (
	"fmt"
	"strings"

	"github.com/stoewer/go-strcase"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
)

// FieldTagOptions are the options for building FieldTag objects.
type FieldTagOptions struct {
	DatabaseKind      string
	DomainName        string
	OutboundName      string
	InboundName       string
	FieldExtensions   *extensions.MikrosFieldExtensions
	MessageExtensions *extensions.MikrosMessageExtensions
}

// FieldTag is the mechanism that allows building and retrieving struct tags from
// a Field for different scenarios.
type FieldTag struct {
	domainTag         string
	outboundTag       string
	outboundFieldName string
	inboundTag        string
}

func newFieldTag(options *FieldTagOptions) *FieldTag {
	var (
		db               = NewTagGenerator(options.DatabaseKind, options.FieldExtensions)
		domainNameMode   = extensions.NamingMode_NAMING_MODE_SNAKE_CASE
		outboundNameMode = extensions.NamingMode_NAMING_MODE_SNAKE_CASE
	)

	if messageDomain := options.MessageExtensions.GetDomain(); messageDomain != nil {
		domainNameMode = messageDomain.GetNamingMode()
	}
	if messageOutbound := options.MessageExtensions.GetOutbound(); messageOutbound != nil {
		outboundNameMode = messageOutbound.GetNamingMode()
	}

	var (
		domainName   = resolveNameForTag(options.DomainName, domainNameMode)
		outboundName = resolveNameForTag(options.OutboundName, outboundNameMode)
	)

	return &FieldTag{
		domainTag:         buildDomainTag(domainName, options.FieldExtensions, db),
		outboundTag:       buildOutboundTag(outboundName, options.FieldExtensions),
		outboundFieldName: outboundName,
		inboundTag:        buildInboundTag(options.InboundName),
	}
}

func resolveNameForTag(fieldName string, mode extensions.NamingMode) string {
	fieldName = strcase.SnakeCase(fieldName)
	if mode == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
		fieldName = strcase.LowerCamelCase(fieldName)
	}

	return fieldName
}

func buildDomainTag(fieldName string, ext *extensions.MikrosFieldExtensions, db TagGenerator) string {
	var (
		tags []*extensions.FieldStructTag
		tag  = "omitempty"
	)

	if domain := ext.GetDomain(); domain != nil {
		tags = domain.GetStructTag()

		if domain.GetAllowEmpty() {
			tag = ""
		}
	}

	return buildTag(fieldName, tag, db.GenerateTag(fieldName), tags)
}

func buildOutboundTag(fieldName string, ext *extensions.MikrosFieldExtensions) string {
	var (
		tags []*extensions.FieldStructTag
		tag  = "omitempty"
	)

	if outbound := ext.GetOutbound(); outbound != nil {
		tags = outbound.GetStructTag()

		if outbound.GetAllowEmpty() {
			tag = ""
		}
	}

	return buildTag(fieldName, tag, "", tags)
}

func buildInboundTag(fieldName string) string {
	return buildTag(fieldName, "", "", nil)
}

func buildTag(fieldName, tag, dbTag string, structTags []*extensions.FieldStructTag) string {
	// Build the base tags
	tags := []string{
		fmt.Sprintf("json:%q", fieldName+prefixComma(tag)),
	}
	if dbTag != "" {
		tags = append(tags, dbTag)
	}

	// Append custom tags from extensions
	for _, st := range structTags {
		tags = append(tags, fmt.Sprintf("%s:%q", st.GetName(), st.GetValue()))
	}

	return "`" + strings.Join(filterEmpty(tags), " ") + "`"
}

func prefixComma(s string) string {
	if s == "" {
		return ""
	}

	return "," + s
}

func filterEmpty(ss []string) []string {
	var out []string
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}

	return out
}

// Domain returns the domain tag for the field.
func (f *FieldTag) Domain() string {
	return f.domainTag
}

// Outbound returns the outbound tag for the field.
func (f *FieldTag) Outbound() string {
	return f.outboundTag
}

// OutboundTagFieldName returns the tag outbound field name for the field.
func (f *FieldTag) OutboundTagFieldName() string {
	return f.outboundFieldName
}

// Inbound returns the inbound tag for the field.
func (f *FieldTag) Inbound() string {
	return f.inboundTag
}
