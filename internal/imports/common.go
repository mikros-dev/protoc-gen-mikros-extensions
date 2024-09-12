package imports

func loadCommonTemplateImports(ctx *Context) []*Import {
	imports := make(map[string]*Import)

	if ctx.HasProtobufValueField {
		imports["protostruct"] = packages["protostruct"]
	}

	if ctx.OutboundHasBitflagField {
		imports["strings"] = packages["strings"]
	}

	return toSlice(imports)
}
