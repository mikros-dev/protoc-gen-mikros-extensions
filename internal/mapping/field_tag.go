package mapping

import (
	"fmt"
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
)

// FieldTagOptions are the options for building FieldTag objects.
type FieldTagOptions struct {
	DatabaseKind    string
	FieldExtensions *extensions.MikrosFieldExtensions
}

// FieldTag is the mechanism that allows building and retrieving struct tags from
// a Field for different scenarios.
type FieldTag struct {
	extensions *extensions.MikrosFieldExtensions
	db         TagGenerator
}

func newFieldTag(options *FieldTagOptions) *FieldTag {
	return &FieldTag{
		extensions: options.FieldExtensions,
		db:         NewTagGenerator(options.DatabaseKind, options.FieldExtensions),
	}
}

// DomainTag returns the domain tag for the field.
func (f *FieldTag) DomainTag(fieldName string) string {
	var (
		domain *extensions.FieldDomainOptions
		tags   []*extensions.FieldDomainStructTag
		tag    = "omitempty"
	)

	if f.extensions != nil {
		domain = f.extensions.GetDomain()
	}

	if domain != nil {
		tags = domain.GetStructTag()

		if domain.GetAllowEmpty() {
			tag = ""
		}
	}

	return f.buildTag(fieldName, tag, f.db.GenerateTag(fieldName), tags)
}

// OutboundTag returns the outbound tag for the field.
func (f *FieldTag) OutboundTag(fieldName string) string {
	var (
		outbound *extensions.FieldOutboundOptions
		tags     []*extensions.FieldDomainStructTag
		tag      = "omitempty"
	)

	if f.extensions != nil {
		outbound = f.extensions.GetOutbound()
	}
	if outbound != nil {
		tags = outbound.GetStructTag()

		if outbound.GetAllowEmpty() {
			tag = ""
		}
	}

	return f.buildTag(fieldName, tag, "", tags)
}

func (f *FieldTag) buildTag(fieldName, tag, dbTab string, structTags []*extensions.FieldDomainStructTag) string {
	// Build the base tags
	tags := []string{
		fmt.Sprintf("json:%q", fieldName+prefixComma(tag)),
	}
	if dbTab != "" {
		tags = append(tags, dbTab)
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

// InboundTag returns the inbound tag for the field.
func (f *FieldTag) InboundTag(fieldName string) string {
	return fmt.Sprintf("`json:\"%s\"`", fieldName)
}
