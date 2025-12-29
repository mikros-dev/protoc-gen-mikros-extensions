package imports

import (
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// Domain represents the 'api/domain.tmpl' importer
type Domain struct{}

// Name returns the template name.
func (d *Domain) Name() spec.Name {
	return spec.NewName("api", "domain")
}

// Load returns a slice of imports for the template
func (d *Domain) Load(ctx *Context, cfg *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	for k, v := range d.loadImportsFromMessages(ctx, cfg, ctx.DomainMessages) {
		imports[k] = v
	}

	return toSlice(imports)
}

func (d *Domain) loadImportsFromMessages(
	ctx *Context,
	cfg *settings.Settings,
	messages []*Message,
) map[string]*Import {
	imports := make(map[string]*Import)

	for _, msg := range messages {
		for _, f := range msg.Fields {
			var (
				call             = cfg.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr) + "("
				conversionToWire = strings.TrimPrefix(f.ConversionDomainToWire, call)
				wireType         = strings.TrimPrefix(f.WireType, "[]*")
			)

			d.addConvertersImport(cfg, conversionToWire, imports)

			if d.addTimeImport(f, imports) {
				continue
			}

			if d.addProtoTimestampImport(wireType, imports) {
				continue
			}

			// Import other modules?
			if module, ok := needsImportAnotherProtoModule(
				conversionToWire,
				wireType,
				ctx.ModuleName,
				msg.Receiver,
			); ok {
				imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
			}
		}
	}

	return imports
}

func (d *Domain) addConvertersImport(cfg *settings.Settings, conversionToWire string, imports map[string]*Import) {
	// Import user converters package?
	if i, ok := needsUserConvertersPackage(cfg, conversionToWire); ok {
		imports["converters"] = i
	}
}

func (d *Domain) addTimeImport(f *Field, imports map[string]*Import) bool {
	// Import time package?
	if f.IsProtobufTimestamp && strings.Contains(f.DomainType, "time.Time") {
		imports["time"] = packages["time"]

		if !f.IsArray {
			// By returning true we signal that no import checking should be
			// done anymore
			return true
		}
	}

	return false
}

func (d *Domain) addProtoTimestampImport(wireType string, imports map[string]*Import) bool {
	if strings.HasPrefix(wireType, "ts.") || strings.HasPrefix(wireType, "*ts.") {
		imports["prototimestamp"] = packages["prototimestamp"]

		// By returning true we signal that no import checking should be
		// done anymore
		return true
	}

	return false
}
