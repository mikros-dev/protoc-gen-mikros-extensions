package imports

func loadCommonTemplateImports(ctx *Context) []*Import {
	ipt := make(map[string]*Import)

	if ctx.HasProtobufValueField {
		ipt["protostruct"] = packages["protostruct"]
	}

	if ctx.OutboundHasBitflagField {
		ipt["strings"] = packages["strings"]
	}

	return toSlice(ipt)
}
