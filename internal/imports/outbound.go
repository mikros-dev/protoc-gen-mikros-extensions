package imports

import (
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// Outbound represents the 'api/outbound.tmpl' importer
type Outbound struct{}

// Name returns the template name.
func (o *Outbound) Name() spec.Name {
	return spec.NewName("api", "outbound")
}

// Load returns a slice of imports for the template.
func (o *Outbound) Load(ctx *Context, cfg *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	for k, v := range o.loadImportsFromMessages(ctx, cfg, ctx.OutboundMessages) {
		imports[k] = v
	}

	return toSlice(imports)
}

func (o *Outbound) loadImportsFromMessages(
	ctx *Context,
	cfg *settings.Settings,
	messages []*Message,
) map[string]*Import {
	imports := make(map[string]*Import)

	for _, msg := range messages {
		for _, f := range msg.Fields {
			o.processField(ctx, cfg, msg.Receiver, f, imports)
		}
	}

	return imports
}

func (o *Outbound) processField(
	ctx *Context,
	cfg *settings.Settings,
	receiver string,
	f *Field,
	imports map[string]*Import,
) {
	if f.OutboundHide {
		return
	}

	if ipt := fieldHasCustomImport(f.ProtoField); ipt != nil {
		imports[ipt.Name] = ipt
		return
	}

	var (
		outboundType   = strings.TrimPrefix(f.OutboundType, "[]*")
		conversionCall = f.ConversionWireOutputToOutbound
	)

	// Import user converters package?
	if i, ok := needsUserConvertersPackage(cfg, conversionCall); ok {
		imports["converters"] = i
	}

	// Import time package?
	if f.IsProtobufTimestamp {
		imports["time"] = packages["time"]
		return
	}

	// Import proto timestamp package?
	if strings.HasPrefix(outboundType, "ts.") {
		imports["prototimestamp"] = packages["prototimestamp"]
		return
	}

	// Import strings?
	if f.IsOutboundBitflag {
		// Is this bitflag from another module?
		if parts := strings.Split(f.ConversionWireOutputToOutbound, ","); len(parts) == 3 {
			valuesVar := parts[1]
			if strings.Contains(valuesVar, ".") {
				module := strings.TrimSpace(strings.Split(valuesVar, ".")[0])
				imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
			}
		}

		return
	}

	// Import other modules?
	if module, ok := needsImportAnotherProtoModule("", outboundType, ctx.ModuleName, receiver); ok {
		imports[module] = importAnotherModule(module, ctx.ModuleName, ctx.FullPath)
		return
	}
}

func fieldHasCustomImport(field *protobuf.Field) *Import {
	if ext := mikros_extensions.LoadFieldExtensions(field.Proto); ext != nil {
		if outbound := ext.GetOutbound(); outbound != nil {
			if ipt := outbound.GetCustomImport(); ipt != nil {
				return &Import{
					Alias: ipt.GetAlias(),
					Name:  ipt.GetName(),
				}
			}
		}
	}

	return nil
}
