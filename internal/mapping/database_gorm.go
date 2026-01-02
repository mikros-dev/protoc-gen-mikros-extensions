package mapping

import (
	"fmt"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
)

type gormGenerator struct {
	defs *extensions.MikrosFieldExtensions
}

// GenerateTag generates a struct tag for the given field name.
func (g *gormGenerator) GenerateTag(_ string) string {
	if g.defs == nil || g.defs.GetDatabase() == nil {
		return ""
	}

	db := g.defs.GetDatabase()
	return buildGormTag(db)
}

func buildGormTag(db *extensions.FieldDatabaseOptions) string {
	var (
		tag   string
		flags = []struct {
			Condition bool
			FlagName  string
		}{
			{
				Condition: db.GetIndex(),
				FlagName:  "index",
			},
			{
				Condition: db.GetUnique(),
				FlagName:  "unique",
			},
			{
				Condition: db.GetUniqueIndex(),
				FlagName:  "uniqueIndex",
			},
			{
				Condition: db.GetPrimaryKey(),
				FlagName:  "primaryKey",
			},
			{
				Condition: db.GetAutoIncrement(),
				FlagName:  "autoIncrement",
			},
		}
	)

	if n := db.GetName(); n != "" {
		tag += "column=" + n
	}

	for _, flag := range flags {
		if flag.Condition {
			if tag != "" {
				tag += ","
			}
			tag += flag.FlagName
		}
	}

	if tag != "" {
		tag = fmt.Sprintf(`gorm:"%s"`, tag)
	}

	return tag
}
