package imports

func loadEnumTemplateImports() []*Import {
	ipt := map[string]*Import{
		"strings": packages["strings"],
	}

	return toSlice(ipt)
}
