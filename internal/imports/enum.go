package imports

func loadEnumTemplateImports() []*Import {
	imports := map[string]*Import{
		"strings": packages["strings"],
	}

	return toSlice(imports)
}
