package converters

import (
	"fmt"

	"github.com/stoewer/go-strcase"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
)

// Database represents a database conversion mechanism used to generate
// the proper source code inside the templates.
type Database struct {
	Kind DatabaseKind
	defs *mikros_extensions.MikrosFieldExtensions
}

// DatabaseKind represents the database kind.
type DatabaseKind int

// Supported database kinds.
const (
	MongoDB DatabaseKind = iota
	Gorm
)

func databaseFromString(kind string, defs *mikros_extensions.MikrosFieldExtensions) *Database {
	db := &Database{
		defs: defs,
	}

	if kind == "mongo" {
		db.Kind = MongoDB
	}
	if kind == "gorm" {
		db.Kind = Gorm
	}

	return db
}

// FieldName returns the database field name for the given name.
func (d *Database) FieldName(name string) string {
	fieldName := name
	if d.Kind == MongoDB {
		if name == "id" {
			fieldName = "_id"
		}
	}

	if d.defs != nil {
		if db := d.defs.GetDatabase(); db != nil {
			if n := db.GetName(); n != "" {
				fieldName = n
			}
		}
	}

	return strcase.SnakeCase(fieldName)
}

// Tag returns the database struct tag for the given name.
func (d *Database) Tag(name string) string {
	if d.Kind == MongoDB {
		return d.handleMongoTag(name)
	}

	if d.Kind == Gorm {
		return d.handleGormTag()
	}

	return ""
}

func (d *Database) handleMongoTag(name string) string {
	omitempty := ",omitempty"
	if d.defs != nil {
		if db := d.defs.GetDatabase(); db != nil {
			if db.GetAllowEmpty() {
				omitempty = ""
			}
		}
	}

	return fmt.Sprintf(`bson:"%s%s"`, d.FieldName(name), omitempty)
}

func (d *Database) handleGormTag() string {
	if d.defs == nil {
		return ""
	}

	if db := d.defs.GetDatabase(); db != nil {
		return buildGormTag(db)
	}

	return ""
}

func buildGormTag(db *mikros_extensions.FieldDatabaseOptions) string {
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
