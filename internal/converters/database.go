package converters

import (
	"fmt"

	"github.com/iancoleman/strcase"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
)

type Database struct {
	Kind DatabaseKind
	defs *extensions.FieldDatabaseOptions
}

type DatabaseKind int

const (
	MongoDB DatabaseKind = iota
)

func databaseFromString(kind string, defs *extensions.FieldDatabaseOptions) *Database {
	db := &Database{
		defs: defs,
	}

	if kind == "mongo" {
		db.Kind = MongoDB
	}

	return db
}

func (d *Database) FieldName(name string) string {
	fieldName := name
	if d.Kind == MongoDB {
		if name == "id" {
			fieldName = "_id"
		}
	}

	if d.defs != nil {
		if n := d.defs.GetName(); n != "" {
			fieldName = n
		}
	}

	return strcase.ToSnake(fieldName)
}

func (d *Database) Tag(name string) string {
	if d.Kind == MongoDB {
		omitempty := ",omitempty"
		if d.defs != nil {
			if d.defs.GetAllowEmpty() {
				omitempty = ""
			}
		}

		return fmt.Sprintf(`bson:"%s%s"`, d.FieldName(name), omitempty)
	}

	return ""
}
