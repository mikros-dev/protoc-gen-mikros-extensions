package imports

import (
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/imports"
)

func loadCommonTemplateImports(ctx *Context) []*imports.Import {
	ipt := make(map[string]*imports.Import)

	if ctx.HasProtobufValueField {
		ipt["protostruct"] = packages["protostruct"]
	}

	if ctx.OutboundHasBitflagField {
		ipt["strings"] = packages["strings"]
	}

	return toSlice(ipt)
}
