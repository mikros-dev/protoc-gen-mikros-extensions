package imports

import (
	"strings"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/settings"
)

func loadTestingTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	ipt := map[string]*Import{
		"math/rand":    packages["math/rand"],
		"reflect":      packages["reflect"],
		ctx.ModuleName: importAnotherModule(ctx.ModuleName, ctx.ModuleName, ctx.FullPath),
	}

	for _, message := range ctx.DomainMessages {
		for _, f := range message.Fields {
			var (
				binding = f.TestingBinding
				call    = f.TestingCall
			)

			if module, ok := needsImportAnotherProtoModule(binding, "", ctx.ModuleName, message.Receiver); ok {
				ipt[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
			}

			if i, ok := needsUserConvertersPackage(cfg, binding); ok {
				ipt["converters"] = i
			}

			if f.IsProtobufTimestamp {
				ipt["time"] = packages["time"]
			}

			if strings.Contains(call, "FromString") {
				module := strings.Split(call, ".")[0]
				ipt[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
				continue
			}

			if module, ok := getModuleFromZeroValueCall(call, f.ProtoField); ok {
				ipt[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
			}
		}
	}

	return toSlice(ipt)
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
		if ('a' <= b && b <= 'z') || ('A' <= b && b <= 'Z') || ('0' <= b && b <= '9') || b == ' ' {
			result.WriteByte(b)
		}
	}

	return result.String()
}