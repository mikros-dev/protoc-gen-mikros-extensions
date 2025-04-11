package types

import (
	"strings"
	"text/template"

	"github.com/stoewer/go-strcase"
)

type Kind int

const (
	KindApi Kind = iota
	KindTest
)

// Validator is a behavior that the templates context and addons must implement
// to validate their execution.
type Validator interface {
	GetTemplateValidator(name Name, ctx interface{}) (ValidateForExecution, bool)
}

type ValidateForExecution func() bool

// HelperApi gives the API available for all templates to be used.
func HelperApi() map[string]interface{} {
	return template.FuncMap{
		"toLowerCamelCase": strcase.LowerCamelCase,
		"firstLower": func(s string) string {
			c := s[0]
			return strings.ToLower(string(c))
		},
		"toSnake":     strcase.SnakeCase,
		"toCamelCase": strcase.UpperCamelCase,
		"toKebab":     strcase.KebabCase,
		"trimSuffix":  strings.TrimSuffix,
	}
}
