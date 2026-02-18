package spec

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
	GetTemplateValidator(name Name, ctx interface{}) (ExecutionFunc, bool)
}

// ExecutionFunc is a function that must return true if the template
// should be executed.
type ExecutionFunc func() bool

// DefaultFuncMap gives the API available for all templates to be used.
func DefaultFuncMap() map[string]interface{} {
	return template.FuncMap{
		"toLowerCamelCase": strcase.LowerCamelCase,
		"firstLower": func(s string) string {
			if len(s) == 0 {
				return ""
			}
			return strings.ToLower(s[:1])
		},
		"toSnake":     strcase.SnakeCase,
		"toCamelCase": strcase.UpperCamelCase,
		"toKebab":     strcase.KebabCase,
		"trimSuffix":  strings.TrimSuffix,
		"toUpper":     strings.ToUpper,
	}
}
