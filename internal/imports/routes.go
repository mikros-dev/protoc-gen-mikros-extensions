package imports

func loadRoutesTemplateImports(ctx *Context) []*Import {
	ipt := map[string]*Import{
		packages["fasthttp"].Name: packages["fasthttp"],
	}

	for _, m := range ctx.Methods {
		if m.HasRequiredBody {
			ipt[packages["errors"].Name] = packages["errors"]
			ipt[packages["json"].Name] = packages["json"]
		}

		if m.HasQueryArguments || m.HasHeaderArguments {
			ipt[packages["fmt"].Name] = packages["fmt"]
		}
	}

	return toSlice(ipt)
}
