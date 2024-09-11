package imports

import (
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/imports"
)

func loadEnumTemplateImports() []*imports.Import {
	ipt := map[string]*imports.Import{
		"strings": packages["strings"],
	}

	return toSlice(ipt)
}
