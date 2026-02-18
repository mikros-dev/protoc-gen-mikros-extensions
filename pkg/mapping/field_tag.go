package mapping

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/stoewer/go-strcase"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
)

// FieldTagOptions are the options for building FieldTag objects.
type FieldTagOptions struct {
	FieldNaming *FieldNaming `validate:"required"`
	*FieldMappingContextOptions
}

// FieldTag is the mechanism that allows building and retrieving struct tags from
// a field for different scenarios.
type FieldTag struct {
	domainTag         string
	outboundTag       string
	outboundFieldName string
	inboundTag        string
}

// NewFieldTag returns a new FieldTag instance.
func NewFieldTag(options *FieldTagOptions) (*FieldTag, error) {
	validate := options.Validate
	if validate == nil {
		validate = validator.New()
	}
	if err := validate.Struct(options); err != nil {
		return nil, err
	}

	var databaseKind string
	if options.Settings != nil {
		databaseKind = options.Settings.Database.Kind
	}

	var (
		fieldExtensions   = loadFieldExtensions(options.ProtoField)
		messageExtensions = loadMessageExtensions(options.ProtoMessage)
		db                = NewTagGenerator(databaseKind, fieldExtensions)
		domainNameMode    = extensions.NamingMode_NAMING_MODE_SNAKE_CASE
		outboundNameMode  = extensions.NamingMode_NAMING_MODE_SNAKE_CASE
	)

	if messageDomain := messageExtensions.GetDomain(); messageDomain != nil {
		domainNameMode = messageDomain.GetNamingMode()
	}
	if messageOutbound := messageExtensions.GetOutbound(); messageOutbound != nil {
		outboundNameMode = messageOutbound.GetNamingMode()
	}

	var (
		domainName   = resolveNameForTag(options.FieldNaming.Domain(), domainNameMode)
		outboundName = resolveNameForTag(options.FieldNaming.Outbound(), outboundNameMode)
	)

	return &FieldTag{
		domainTag:         buildDomainTag(domainName, fieldExtensions, db),
		outboundTag:       buildOutboundTag(outboundName, fieldExtensions),
		outboundFieldName: outboundName,
		inboundTag:        buildInboundTag(options.FieldNaming.Inbound()),
	}, nil
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
