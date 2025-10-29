package imports

import (
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

func loadWireTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	imports := map[string]*Import{
		packages["time"].Name:           packages["time"],
		packages["prototimestamp"].Name: packages["prototimestamp"],
		packages["protostruct"].Name:    packages["protostruct"],
	}

	for k, v := range loadImportsFromMessagesToWire(ctx, cfg, ctx.DomainMessages) {
		imports[k] = v
	}

	return toSlice(imports)
}

func loadImportsFromMessagesToWire(ctx *Context, cfg *settings.Settings, messages []*Message) map[string]*Import {
	imports := make(map[string]*Import)

	for _, msg := range messages {
		for _, f := range msg.Fields {
			if f.IsMessage && !f.IsArray {
				// Don't need to check non array messages because they only
				// call IntoDomain method.
				continue
			}

			var (
				conversionToDomain = f.ConversionWireToDomain
				domainType         = strings.TrimPrefix(f.DomainType, "[]*")
			)

			// Import user converters package?
			if i, ok := needsUserConvertersPackage(cfg, conversionToDomain); ok {
				imports["converters"] = i
			}

			// Import time package?
			if f.IsProtobufTimestamp {
				imports["time"] = packages["time"]

				if !f.IsArray {
					continue
				}
			}

			// Import proto timestamp package?
			if strings.HasPrefix(domainType, "ts.") || strings.HasPrefix(domainType, "*ts.") {
				imports["prototimestamp"] = packages["prototimestamp"]
				continue
			}

			// Import other modules?
			if module, ok := needsImportAnotherProtoModule(
				conversionToDomain,
				domainType,
				ctx.ModuleName,
				msg.Receiver,
			); ok {
				imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
			}
		}
	}

	return imports
}
