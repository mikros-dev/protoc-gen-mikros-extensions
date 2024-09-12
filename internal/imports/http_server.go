package imports

func loadHttpServerTemplateImports() []*Import {
	ipt := map[string]*Import{
		packages["context"].Name:         packages["context"],
		packages["errors"].Name:          packages["errors"],
		packages["fasthttp"].Name:        packages["fasthttp"],
		packages["fasthttp-router"].Name: packages["fasthttp-router"],
	}

	return toSlice(ipt)
}

func loadTestingHttpServerTemplateImports(ctx *Context) []*Import {
	ipt := map[string]*Import{
		packages["fasthttp-router"].Name: packages["fasthttp-router"],
		ctx.ModuleName:                   importAnotherModule(ctx.ModuleName, ctx.ModuleName, ctx.FullPath),
	}

	return toSlice(ipt)
}
