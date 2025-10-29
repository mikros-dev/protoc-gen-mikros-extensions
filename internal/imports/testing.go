package imports

import (
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

func loadTestingTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	var (
		importTestingRule = false
		imports           = map[string]*Import{
			"math/rand":    packages["math/rand"],
			"reflect":      packages["reflect"],
			ctx.ModuleName: importAnotherModule(ctx.ModuleName, ctx.ModuleName, ctx.FullPath),
		}
	)

	for _, message := range ctx.DomainMessages {
		for _, f := range message.Fields {
			var (
				binding   = f.TestingBinding
				call      = buildTestingCall(cfg, f.TestingCall)
				fieldType = f.DomainType
			)

			if hasTestingOption(f) {
				importTestingRule = true
			}

			addModuleIfNeeded(imports, binding, fieldType, ctx.ModuleName, message.Receiver, ctx.FullPath)
			addConvertersIfNeeded(imports, cfg, binding)
			addTimeIfNeeded(imports, f)

			if handled := handleFromStringCall(imports, call, ctx); handled {
				continue
			}

			addModuleIfNeededFromZeroValue(imports, call, f.ProtoField, ctx)
		}
	}

	addTestingRuleImportIfNeeded(imports, importTestingRule, cfg)

	return toSlice(imports)
}

func buildTestingCall(cfg *settings.Settings, testingCall string) string {
	cfgCall := cfg.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr) + "("
	return strings.TrimPrefix(testingCall, cfgCall)
}

func hasTestingOption(f *Field) bool {
	options := mikros_extensions.LoadFieldExtensions(f.ProtoField.Proto)
	return options != nil && options.GetTesting() != nil
}

func handleFromStringCall(imports map[string]*Import, call string, ctx *Context) bool {
	if strings.Contains(call, "FromString") {
		module := strings.Split(call, ".")[0]
		imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
		return true
	}

	return false
}

func addModuleIfNeededFromZeroValue(imports map[string]*Import, call string, field *protobuf.Field, ctx *Context) {
	if module, ok := getModuleFromZeroValueCall(call, field); ok {
		imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
	}
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
			_ = result.WriteByte(b)
		}
	}

	return result.String()
}

func addTestingRuleImportIfNeeded(imports map[string]*Import, importTestingRule bool, cfg *settings.Settings) {
	if importTestingRule && cfg.Testing != nil && cfg.Testing.PackageImport != nil {
		imports[cfg.Testing.PackageImport.Name] = &Import{
			Name:  cfg.Testing.PackageImport.Name,
			Alias: cfg.Testing.PackageImport.Alias,
		}
	}
}
