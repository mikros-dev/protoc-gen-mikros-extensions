package types

import (
	"strings"
	"text/template"

	"github.com/stoewer/go-strcase"
)

// Kind represents the kind of template.
type Kind int

// Supported template kinds.
const (
	KindAPI Kind = iota
	KindTest
)

// Validator is a behavior that the templates' contexts and addons must implement
// to validate their execution.
type Validator interface {
	GetTemplateValidator(name Name, ctx interface{}) (ValidateForExecution, bool)
}

// ValidateForExecution is a function that must return true if the template
// should be executed.
type ValidateForExecution func() bool

// HelperAPI gives the API available for all templates to be used.
func HelperAPI() map[string]interface{} {
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
		"toUpper":     strings.ToUpper,
	}
}
