package template

import (
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

type Kind int

const (
	KindApi Kind = iota
	KindTest
)

// Validator is a behavior that the templates context and addons must implement
// to validate their execution.
type Validator interface {
	GetTemplateValidator(ctx interface{}, name Name) (ValidateForExecution, bool)
}

type ValidateForExecution func() bool

// HelperApi gives the API available for all templates to be used.
func HelperApi() map[string]interface{} {
	return template.FuncMap{
		"toLowerCamelCase": strcase.ToLowerCamel,
		"firstLower": func(s string) string {
			c := s[0]
			return strings.ToLower(string(c))
		},
		"toSnake":     strcase.ToSnake,
		"toCamelCase": strcase.ToCamel,
		"toKebab":     strcase.ToKebab,
		"trimSuffix":  strings.TrimSuffix,
	}
}
