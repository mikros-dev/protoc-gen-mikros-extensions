package mapping

import (
	"fmt"

	"github.com/stoewer/go-strcase"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
)

type mongoGenerator struct {
	defs *extensions.MikrosFieldExtensions
}

func (g *mongoGenerator) fieldName(name string) string {
	fieldName := name
	if name == "id" {
		fieldName = "_id"
	}

	if g.defs != nil {
		if db := g.defs.GetDatabase(); db != nil {
			if n := db.GetName(); n != "" {
				fieldName = n
			}
		}
	}

	return strcase.SnakeCase(fieldName)
}

// GenerateTag generates a struct tag for the given field name.
func (g *mongoGenerator) GenerateTag(name string) string {
	omitempty := ",omitempty"
	if g.defs != nil {
		if db := g.defs.GetDatabase(); db != nil {
			if db.GetAllowEmpty() {
				omitempty = ""
			}
		}
	}

	return fmt.Sprintf(`bson:"%s%s"`, g.fieldName(name), omitempty)
}
