package imports

import (
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// Testing represents the 'testing/testing.tmpl' importer
type Testing struct{}

// Name returns the template name.
func (t *Testing) Name() spec.Name {
	return spec.NewName("testing", "testing")
}

// Load returns a slice of imports for the template.
func (t *Testing) Load(ctx *Context, cfg *settings.Settings) []*Import {
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
				call      = t.buildTestingCall(cfg, f.TestingCall)
				fieldType = f.DomainType
			)

			if t.hasTestingOption(f) {
				importTestingRule = true
			}

			addModuleIfNeeded(imports, binding, fieldType, ctx.ModuleName, message.Receiver, ctx.FullPath)
			addConvertersIfNeeded(imports, cfg, binding)
			addTimeIfNeeded(imports, f)

			if handled := t.handleFromStringCall(imports, call, ctx); handled {
				continue
			}

			t.addModuleIfNeededFromZeroValue(imports, call, f.ProtoField, ctx)
		}
	}

	t.addTestingRuleImportIfNeeded(imports, importTestingRule, cfg)

	return toSlice(imports)
}

func (t *Testing) buildTestingCall(cfg *settings.Settings, testingCall string) string {
	cfgCall := cfg.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr) + "("
	return strings.TrimPrefix(testingCall, cfgCall)
}

func (t *Testing) hasTestingOption(f *Field) bool {
	options := extensions.LoadFieldExtensions(f.ProtoField.Proto)
	return options != nil && options.GetTesting() != nil
}

func (t *Testing) handleFromStringCall(imports map[string]*Import, call string, ctx *Context) bool {
	if strings.Contains(call, "FromString") {
		module := strings.Split(call, ".")[0]
		imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
		return true
	}

	return false
}

func (t *Testing) addModuleIfNeededFromZeroValue(
	imports map[string]*Import,
	call string,
	field *protobuf.Field,
	ctx *Context,
) {
	if module, ok := t.getModuleFromZeroValueCall(call, field); ok {
		imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
	}
}

func (t *Testing) getModuleFromZeroValueCall(call string, field *protobuf.Field) (string, bool) {
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

func (t *Testing) addTestingRuleImportIfNeeded(
	imports map[string]*Import,
	importTestingRule bool,
	cfg *settings.Settings,
) {
	if importTestingRule && cfg.Testing != nil && cfg.Testing.PackageImport != nil {
		imports[cfg.Testing.PackageImport.Name] = &Import{
			Name:  cfg.Testing.PackageImport.Name,
			Alias: cfg.Testing.PackageImport.Alias,
		}
	}
}
