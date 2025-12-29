package imports

import (
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// WireInput represents the 'api/wire_input.tmpl' importer
type WireInput struct{}

// Name returns the template name.
func (w *WireInput) Name() spec.Name {
	return spec.NewName("api", "wire_input")
}

// Load returns a slice of imports for the template.
func (w *WireInput) Load(ctx *Context, cfg *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	for k, v := range w.loadImportsFromMessages(ctx, cfg, ctx.WireInputMessages) {
		imports[k] = v
	}

	return toSlice(imports)
}

func (w *WireInput) loadImportsFromMessages(
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

			w.addUserConvertersImport(imports, cfg, conversionToWire)
			w.addTimeImportIfNeeded(imports, f, msg)

			// Skip non-array protobuf timestamps after adding time import (if needed)
			if f.IsProtobufTimestamp && !f.IsArray {
				continue
			}

			if w.addProtoTimestampImportIfNeeded(imports, wireType) {
				continue
			}

			w.addOtherModuleImportIfNeeded(imports, conversionToWire, wireType, msg, ctx)
		}
	}

	return imports
}

// addUserConvertersImport adds the user converters package if it's needed.
func (w *WireInput) addUserConvertersImport(
	imports map[string]*Import,
	cfg *settings.Settings,
	conversionToWire string,
) {
	if i, ok := needsUserConvertersPackage(cfg, conversionToWire); ok {
		imports["converters"] = i
	}
}

// addTimeImportIfNeeded imports time when handling wire input protobuf timestamps mapped to time.Time.
func (w *WireInput) addTimeImportIfNeeded(
	imports map[string]*Import,
	f *Field,
	msg *Message,
) {
	if f.IsProtobufTimestamp && msg.IsWireInputKind && strings.Contains(f.DomainType, "time.Time") {
		imports["time"] = packages["time"]
	}
}

// addProtoTimestampImportIfNeeded imports prototimestamp when wire type references ts.Timestamp.
func (w *WireInput) addProtoTimestampImportIfNeeded(imports map[string]*Import, wireType string) bool {
	if strings.HasPrefix(wireType, "ts.") || strings.HasPrefix(wireType, "*ts.") {
		imports["prototimestamp"] = packages["prototimestamp"]
		return true
	}
	return false
}

// addOtherModuleImportIfNeeded imports another proto module if required by conversion or wire type.
func (w *WireInput) addOtherModuleImportIfNeeded(
	imports map[string]*Import,
	conversionToWire, wireType string,
	msg *Message,
	ctx *Context,
) {
	if module, ok := needsImportAnotherProtoModule(
		conversionToWire,
		wireType,
		ctx.ModuleName,
		msg.Receiver,
	); ok {
		imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
	}
}
