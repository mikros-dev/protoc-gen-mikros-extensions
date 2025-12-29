package converters

import (
	"strings"

	"github.com/stoewer/go-strcase"
)

// TrimPackageName trims the package name from the given name.
func TrimPackageName(name, packageName string) string {
	if name == "" {
		return ""
	}

	return strings.TrimPrefix(name, "."+packageName+".")
}

func inboundOutboundCamelCase(s string) string {
	return strcase.LowerCamelCase(s)
}
