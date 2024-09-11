package context

import (
	"strings"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/settings"
)

func loadOutboundTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	for k, v := range loadOutboundImportsFromMessages(ctx, cfg, ctx.OutboundMessages()) {
		imports[k] = v
	}

	return toSlice(imports)
}

func loadOutboundImportsFromMessages(ctx *Context, cfg *settings.Settings, messages []*Message) map[string]*Import {
	imports := make(map[string]*Import)

	for _, msg := range messages {
		for _, f := range msg.Fields {
			var (
				outboundType   = strings.TrimPrefix(f.OutboundType(), "[]*")
				conversionCall = f.ConvertWireOutputToOutbound("r")
			)

			// Import user converters package?
			if i, ok := needsUserConvertersPackage(cfg, conversionCall); ok {
				imports["converters"] = i
			}

			// Import time package?
			if f.field.IsTimestamp() {
				imports["time"] = packages["time"]
				continue
			}

			// Import proto timestamp package?
			if strings.HasPrefix(outboundType, "ts.") {
				imports["prototimestamp"] = packages["prototimestamp"]
				continue
			}

			// Import strings?
			if f.IsOutboundBitflag() {
				// Is this bitflag from another module?
				if parts := strings.Split(f.ConvertWireOutputToOutbound("r"), ","); len(parts) == 3 {
					valuesVar := parts[1]
					if strings.Contains(valuesVar, ".") {
						module := strings.TrimSpace(strings.Split(valuesVar, ".")[0])
						imports[module] = importAnotherModule(module, ctx.pkg.ModuleName, ctx.pkg.FullPath)
					}
				}

				continue
			}

			// Import other modules?
			if module, ok := needsImportAnotherProtoModule("", outboundType, ctx.ModuleName, f.messageReceiver); ok {
				imports[module] = importAnotherModule(module, ctx.pkg.ModuleName, ctx.pkg.FullPath)
				continue
			}
		}
	}

	return imports
}
