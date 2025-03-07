package converters

import (
	"fmt"

	"github.com/iancoleman/strcase"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
)

type Database struct {
	Kind DatabaseKind
	defs *mikros_extensions.MikrosFieldExtensions
}

type DatabaseKind int

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

	return strcase.ToSnake(fieldName)
}

func (d *Database) Tag(name string) string {
	if d.Kind == MongoDB {
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

	if d.Kind == Gorm {
		if d.defs != nil {
			if db := d.defs.GetDatabase(); db != nil {
				return getPostgresTag(db)
			}
		}
	}

	return ""
}

func getPostgresTag(db *mikros_extensions.FieldDatabaseOptions) string {
	var tag string

	if n := db.GetName(); n != "" {
		tag += "column=" + n
	}

	if db.GetIndex() {
		if tag != "" {
			tag += ","
		}

		tag += "index"
	}

	if db.GetUnique() {
		if tag != "" {
			tag += ","
		}

		tag += "unique"
	}

	if db.GetUniqueIndex() {
		if tag != "" {
			tag += ","
		}

		tag += "uniqueIndex"
	}

	if db.GetPrimaryKey() {
		if tag != "" {
			tag += ","
		}

		tag += "primaryKey"
	}

	if db.GetAutoIncrement() {
		if tag != "" {
			tag += ","
		}

		tag += "autoIncrement"
	}

	if tag != "" {
		tag = fmt.Sprintf(`gorm:"%s"`, tag)
	}

	return tag
}
