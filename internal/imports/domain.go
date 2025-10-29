package imports

import (
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

func loadDomainTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	for k, v := range loadDomainImportsFromMessages(ctx, cfg, ctx.DomainMessages) {
		imports[k] = v
	}

	return toSlice(imports)
}

func loadDomainImportsFromMessages(ctx *Context, cfg *settings.Settings, messages []*Message) map[string]*Import {
	imports := make(map[string]*Import)

	for _, msg := range messages {
		for _, f := range msg.Fields {
			var (
				call             = cfg.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr) + "("
				conversionToWire = strings.TrimPrefix(f.ConversionDomainToWire, call)
				wireType         = strings.TrimPrefix(f.WireType, "[]*")
			)

			// Import user converters package?
			if i, ok := needsUserConvertersPackage(cfg, conversionToWire); ok {
				imports["converters"] = i
			}

			// Import time package?
			if f.IsProtobufTimestamp && strings.Contains(f.DomainType, "time.Time") {
				imports["time"] = packages["time"]

				if !f.IsArray {
					continue
				}
			}

			// Import proto timestamp package?
			if strings.HasPrefix(wireType, "ts.") || strings.HasPrefix(wireType, "*ts.") {
				imports["prototimestamp"] = packages["prototimestamp"]
				continue
			}

			// Import other modules?
			if module, ok := needsImportAnotherProtoModule(conversionToWire, wireType, ctx.ModuleName, msg.Receiver); ok {
				imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
			}
		}
	}

	return imports
}
