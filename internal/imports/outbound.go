package imports

import (
	"strings"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/settings"
)

func loadOutboundTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	ipt := make(map[string]*Import)

	for k, v := range loadOutboundImportsFromMessages(ctx, cfg, ctx.OutboundMessages) {
		ipt[k] = v
	}

	return toSlice(ipt)
}

func loadOutboundImportsFromMessages(ctx *Context, cfg *settings.Settings, messages []*Message) map[string]*Import {
	ipt := make(map[string]*Import)

	for _, msg := range messages {
		for _, f := range msg.Fields {
			var (
				outboundType   = strings.TrimPrefix(f.OutboundType, "[]*")
				conversionCall = f.ConversionWireOutputToOutbound
			)

			// Import user converters package?
			if i, ok := needsUserConvertersPackage(cfg, conversionCall); ok {
				ipt["converters"] = i
			}

			// Import time package?
			if f.IsProtobufTimestamp {
				ipt["time"] = packages["time"]
				continue
			}

			// Import proto timestamp package?
			if strings.HasPrefix(outboundType, "ts.") {
				ipt["prototimestamp"] = packages["prototimestamp"]
				continue
			}

			// Import strings?
			if f.IsOutboundBitflag {
				// Is this bitflag from another module?
				if parts := strings.Split(f.ConversionWireOutputToOutbound, ","); len(parts) == 3 {
					valuesVar := parts[1]
					if strings.Contains(valuesVar, ".") {
						module := strings.TrimSpace(strings.Split(valuesVar, ".")[0])
						ipt[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
					}
				}

				continue
			}

			// Import other modules?
			if module, ok := needsImportAnotherProtoModule("", outboundType, ctx.ModuleName, msg.Receiver); ok {
				ipt[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
				continue
			}
		}
	}

	return ipt
}
