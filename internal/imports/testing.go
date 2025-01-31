package imports

import (
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

func loadTestingTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	imports := map[string]*Import{
		"math/rand":    packages["math/rand"],
		"reflect":      packages["reflect"],
		ctx.ModuleName: importAnotherModule(ctx.ModuleName, ctx.ModuleName, ctx.FullPath),
	}

	for _, message := range ctx.DomainMessages {
		for _, f := range message.Fields {
			var (
				binding   = f.TestingBinding
				cfgCall   = cfg.GetCommonCall(settings.CommonApiConverters, settings.CommonCallToPtr) + "("
				call      = strings.TrimPrefix(f.TestingCall, cfgCall)
				fieldType = f.DomainType
			)

			if module, ok := needsImportAnotherProtoModule(binding, fieldType, ctx.ModuleName, message.Receiver); ok {
				imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
			}

			if i, ok := needsUserConvertersPackage(cfg, binding); ok {
				imports["converters"] = i
			}

			if f.IsProtobufTimestamp {
				imports["time"] = packages["time"]
			}

			if strings.Contains(call, "FromString") {
				module := strings.Split(call, ".")[0]
				imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
				continue
			}

			if module, ok := getModuleFromZeroValueCall(call, f.ProtoField); ok {
				imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
			}
		}
	}

	return toSlice(imports)
}

func getModuleFromZeroValueCall(call string, field *protobuf.Field) (string, bool) {
	if !strings.Contains(call, "zeroValue") || field.IsTimestamp() || field.IsMap() {
		return "", false
	}

	parts := strings.Split(call, ".")
	if len(parts) != 5 {
		return "", false
	}

	return stripNonAlpha(parts[len(parts)-2]), true
}

func stripNonAlpha(s string) string {
	var result strings.Builder

	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') || ('0' <= b && b <= '9') || b == ' ' || b == '_' {
			result.WriteByte(b)
		}
	}

	return result.String()
}
