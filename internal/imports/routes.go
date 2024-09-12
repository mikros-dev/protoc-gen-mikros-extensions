package imports

func loadRoutesTemplateImports(ctx *Context) []*Import {
	imports := map[string]*Import{
		packages["fasthttp"].Name: packages["fasthttp"],
	}

	for _, m := range ctx.Methods {
		if m.HasRequiredBody {
			imports[packages["errors"].Name] = packages["errors"]
			imports[packages["json"].Name] = packages["json"]
		}

		if m.HasQueryArguments || m.HasHeaderArguments {
			imports[packages["fmt"].Name] = packages["fmt"]
		}
	}

	return toSlice(imports)
}
