package imports

func loadCommonTemplateImports(ctx *Context) []*Import {
	imports := make(map[string]*Import)

	if ctx.OutboundHasBitflagField {
		imports["strings"] = packages["strings"]
	}

	if ctx.UseCommonConverters {
		imports["time"] = packages["time"]
		imports["protostruct"] = packages["protostruct"]
		imports["prototimestamp"] = packages["prototimestamp"]
	}

	return toSlice(imports)
}
