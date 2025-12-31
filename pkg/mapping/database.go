package mapping

import (
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
)

// TagGenerator defines the contract for generating database-specific struct tags.
type TagGenerator interface {
	// GenerateTag generates a struct tag for the given field name.
	GenerateTag(fieldName string) string
}

// NewTagGenerator returns the appropriate generator based on the configuration.
func NewTagGenerator(kind string, defs *extensions.MikrosFieldExtensions) TagGenerator {
	switch kind {
	case "mongo":
		return &mongoGenerator{defs: defs}
	case "gorm":
		return &gormGenerator{defs: defs}
	default:
		return &noopGenerator{}
	}
}

// noopGenerator is a generator that does nothing used when no database
// is specified.
type noopGenerator struct{}

// GenerateTag generates a struct tag for the given field name.
func (n *noopGenerator) GenerateTag(string) string {
	return ""
}
