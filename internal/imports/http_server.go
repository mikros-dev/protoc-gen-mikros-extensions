package imports

func loadHttpServerTemplateImports() []*Import {
	imports := map[string]*Import{
		packages["context"].Name:         packages["context"],
		packages["errors"].Name:          packages["errors"],
		packages["fasthttp"].Name:        packages["fasthttp"],
		packages["fasthttp-router"].Name: packages["fasthttp-router"],
	}

	return toSlice(imports)
}

func loadTestingHttpServerTemplateImports(ctx *Context) []*Import {
	imports := map[string]*Import{
		packages["fasthttp-router"].Name: packages["fasthttp-router"],
		ctx.ModuleName:                   importAnotherModule(ctx.ModuleName, ctx.ModuleName, ctx.FullPath),
	}

	return toSlice(imports)
}
