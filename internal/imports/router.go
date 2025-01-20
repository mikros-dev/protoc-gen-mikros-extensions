package imports

import (
	"fmt"
	"sort"
	"strings"
)

func loadRustRouterTemplateImports(ctx *Context) []*Import {
	imports := map[string]*Import{
		"arc":           {Alias: "std::sync::Arc"},
		"axum":          {Alias: "mikros::axum"},
		"axum_json":     {Alias: "axum::Json"},
		"extension":     {Alias: "axum::extract::{Extension, State}"},
		"methods":       {Alias: fmt.Sprintf("axum::routing::{%s}", strings.ToLower(getHTTPMethods(ctx)))},
		"errors":        {Alias: "mikros::{http::ServiceState, errors as merrors}"},
		"serde_derive":  {Alias: "mikros::serde_derive"},
		"serializer":    {Alias: "serde_derive::{Deserialize, Serialize}"},
		"tonic_request": {Alias: "tonic::Request"},
	}

	if hasQueryArguments(ctx) {
		imports["query"] = &Import{Alias: "axum::extract::Query"}
	}
	if hasPathArguments(ctx) {
		imports["path"] = &Import{Alias: "axum::extract::Path"}
	}
	if hasHeaderArguments(ctx) {
		imports["header_map"] = &Import{Alias: "axum::http::header::HeaderMap"}
	}

	return toSlice(imports)
}

func getHTTPMethods(ctx *Context) string {
	methods := make(map[string]bool)
	for _, m := range ctx.Methods {
		methods[m.HTTPMethod] = true
	}

	var s []string
	for k := range methods {
		s = append(s, k)
	}

	sort.Strings(s)
	return strings.Join(s, ",")
}

func hasHeaderArguments(ctx *Context) bool {
	for _, m := range ctx.Methods {
		if m.HasHeaderArguments {
			return true
		}
	}

	return false
}

func hasQueryArguments(ctx *Context) bool {
	for _, m := range ctx.Methods {
		if m.HasQueryArguments {
			return true
		}
	}

	return false
}

func hasPathArguments(ctx *Context) bool {
	for _, m := range ctx.Methods {
		if m.HasPathArguments {
			return true
		}
	}

	return false
}
