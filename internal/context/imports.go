package context

import (
	"cmp"
	"regexp"
	"slices"
	"strings"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/settings"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/template"
)

var (
	mapTypeRe = regexp.MustCompile(`^map\[\S+]`)
)

// packages represents a list of common packages that can be imported by several
// templates.
var packages = map[string]*Import{
	"context": {
		Name: "context",
	},
	"errors": {
		Name: "errors",
	},
	"fmt": {
		Name: "fmt",
	},
	"json": {
		Name: "encoding/json",
	},
	"math/rand": {
		Name: "math/rand",
	},
	"reflect": {
		Name: "reflect",
	},
	"regex": {
		Name: "regexp",
	},
	"strings": {
		Name: "strings",
	},
	"time": {
		Name: "time",
	},
	"prototimestamp": {
		Name:  "google.golang.org/protobuf/types/known/timestamppb",
		Alias: "ts",
	},
	"protostruct": {
		Name: "google.golang.org/protobuf/types/known/structpb",
	},
	"fasthttp": {
		Name: "github.com/valyala/fasthttp",
	},
	"fasthttp-router": {
		Name: "github.com/fasthttp/router",
	},
	"validation": {
		Name: "github.com/go-ozzo/ozzo-validation/v4",
	},
}

type Import struct {
	Alias string
	Name  string
}

func LoadTemplateImports(ctx *Context, cfg *settings.Settings) map[template.Name][]*Import {
	return map[template.Name][]*Import{
		template.NewName("api", "domain"):          loadDomainTemplateImports(ctx, cfg),
		template.NewName("api", "enum"):            loadEnumTemplateImports(),
		template.NewName("api", "wire"):            loadWireTemplateImports(ctx),
		template.NewName("api", "http_server"):     loadHttpServerTemplateImports(),
		template.NewName("api", "routes"):          loadRoutesTemplateImports(ctx),
		template.NewName("api", "wire_input"):      loadWireInputTemplateImports(ctx, cfg),
		template.NewName("api", "outbound"):        loadOutboundTemplateImports(ctx, cfg),
		template.NewName("api", "common"):          loadCommonTemplateImports(ctx),
		template.NewName("api", "validation"):      loadValidationTemplateImports(ctx, cfg),
		template.NewName("testing", "testing"):     loadTestingTemplateImports(ctx, cfg),
		template.NewName("testing", "http_server"): loadTestingHttpServerTemplateImports(ctx),
	}
}

func toSlice(imports map[string]*Import) []*Import {
	var (
		s     = make([]*Import, len(imports))
		index = 0
	)

	for _, i := range imports {
		s[index] = i
		index += 1
	}

	slices.SortFunc(s, func(a, b *Import) int {
		if a.Alias != "" && b.Alias != "" {
			return cmp.Compare(a.Alias, b.Alias)
		}
		if a.Alias != "" && b.Alias == "" {
			return cmp.Compare(a.Alias, b.Name)
		}
		if a.Alias == "" && b.Alias != "" {
			return cmp.Compare(a.Name, b.Alias)
		}

		return cmp.Compare(a.Name, b.Name)
	})

	return s
}

func loadImportsFromMessages(ctx *Context, cfg *settings.Settings, messages []*Message) map[string]*Import {
	imports := make(map[string]*Import)

	for _, msg := range messages {
		for _, f := range msg.Fields {
			var (
				conversionToWire = f.ConvertDomainTypeToWireType()
				wireType         = strings.TrimPrefix(f.WireType(), "[]*")
			)

			// Import user converters package?
			if i, ok := needsUserConvertersPackage(cfg, conversionToWire); ok {
				imports["converters"] = i
			}

			// Import time package?
			if f.field.IsTimestamp() {
				imports["time"] = packages["time"]
			}

			// Import proto timestamp package?
			if strings.HasPrefix(wireType, "ts.") || strings.HasPrefix(wireType, "*ts.") {
				imports["prototimestamp"] = packages["prototimestamp"]
				continue
			}

			// Import other modules?
			if module, ok := needsImportAnotherProtoModule(conversionToWire, wireType, ctx.ModuleName, f.messageReceiver); ok {
				imports[module] = importAnotherModule(module, ctx.pkg.ModuleName, ctx.pkg.FullPath)
			}
		}
	}

	return imports
}

func needsUserConvertersPackage(cfg *settings.Settings, conversionCall string) (*Import, bool) {
	if dep, ok := cfg.Dependencies["converters"]; ok {
		prefix := cfg.GetDependencyModuleName("converters")
		if strings.HasPrefix(conversionCall, prefix) {
			return &Import{
				Alias: dep.Alias,
				Name:  dep.Import,
			}, true
		}
	}

	return nil, false
}

// needsImportAnotherProtoModule checks if a conversion call that is being made must have
// another module imported.
func needsImportAnotherProtoModule(conversionCall, fieldType, moduleName, receiver string) (string, bool) {
	if m, ok := checkImportNeededFromConversionCall(conversionCall, moduleName, receiver); ok {
		return m, ok
	}

	if m, ok := checkImportNeededFromFieldType(fieldType); ok {
		return m, ok
	}

	return "", false
}

func checkImportNeededFromConversionCall(conversionCall, moduleName, receiver string) (string, bool) {
	// Don't bother checking if it is not a function call
	if !strings.Contains(conversionCall, "(") {
		return "", false
	}

	var (
		parts          = strings.Split(conversionCall, ".")
		ignoredModules = []string{moduleName, receiver, "converters"}
	)

	// The conversion should have a module as prefix, like "something.", and
	// it should split to more than 5 parts because the module name usually
	// repeat in it.
	if len(parts) == 0 || len(parts) < 5 || slices.Contains(ignoredModules, parts[0]) {
		return "", false
	}

	return parts[0], true
}

func checkImportNeededFromFieldType(fieldType string) (string, bool) {
	if strings.HasPrefix(fieldType, "map[") {
		fieldType = mapTypeRe.ReplaceAllString(fieldType, "")
	}

	fieldType = strings.TrimPrefix(fieldType, "*")
	parts := strings.Split(fieldType, ".")

	// fieldType must have two parts here 'module.Type' to require another
	// module to be imported
	if len(parts) == 2 {
		return parts[0], true
	}

	return "", false
}

func importAnotherModule(moduleName, currentModuleName, importPath string) *Import {
	return &Import{
		Alias: moduleName,
		Name:  strings.ReplaceAll(importPath, currentModuleName, moduleName),
	}
}
