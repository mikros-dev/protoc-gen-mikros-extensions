package context

import (
	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/imports"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template"
)

type templateImport struct {
	Alias string
	Name  string
}

func loadImports(ctx *Context, cfg *settings.Settings) map[template.Name][]*templateImport {
	var (
		tplImports = imports.LoadTemplateImports(toImportsContext(ctx), cfg)
		ctxImport  = make(map[template.Name][]*templateImport)
	)

	for k, ipt := range tplImports {
		v := make([]*templateImport, len(ipt))
		for i, ii := range ipt {
			v[i] = &templateImport{
				Alias: ii.Alias,
				Name:  ii.Name,
			}
		}
		ctxImport[k] = v
	}

	return ctxImport
}

func toImportsContext(ctx *Context) *imports.Context {
	var (
		methods        []*imports.Method
		domain         []*imports.Message
		outbound       []*imports.Message
		validate       []*imports.Message
		wireExtensions []*imports.Message
		wireInput      []*imports.Message
	)

	fieldToImportField := func(f *Field) *imports.Field {
		return &imports.Field{
			IsArray:                        f.IsArray,
			IsProtobufTimestamp:            f.ProtoField.IsTimestamp(),
			IsOutboundBitflag:              f.IsOutboundBitflag(),
			ConversionDomainToWire:         f.ConvertDomainTypeToWireType(),
			ConversionWireOutputToOutbound: f.ConvertWireOutputToOutbound("r"),
			WireType:                       f.WireType(),
			OutboundType:                   f.OutboundType(),
			TestingBinding:                 f.TestingValueBinding(),
			TestingCall:                    f.TestingValueCall(),
			ValidationCall:                 f.ValidationCall(),
			ProtoField:                     f.ProtoField,
		}
	}

	messageToImportMessage := func(m *Message) *imports.Message {
		var fields []*imports.Field
		for _, f := range m.Fields {
			fields = append(fields, fieldToImportField(f))
		}

		return &imports.Message{
			ValidationNeedsCustomRuleOptions: m.ValidationNeedsCustomRuleOptions(),
			Receiver:                         m.receiver,
			Fields:                           fields,
			ProtoMessage:                     m.ProtoMessage,
		}
	}

	for _, m := range ctx.Methods {
		methods = append(methods, &imports.Method{
			HasRequiredBody:    m.HasRequiredBody(),
			HasQueryArguments:  m.HasHeaderArguments(),
			HasHeaderArguments: m.HasQueryArguments(),
		})
	}

	for _, m := range ctx.DomainMessages() {
		domain = append(domain, messageToImportMessage(m))
	}

	for _, m := range ctx.OutboundMessages() {
		outbound = append(outbound, messageToImportMessage(m))
	}

	for _, m := range ctx.WireInputMessages() {
		wireInput = append(wireInput, messageToImportMessage(m))
	}

	for _, m := range ctx.WireExtensions() {
		wireExtensions = append(wireExtensions, messageToImportMessage(m))
	}

	for _, m := range ctx.ValidatableMessages() {
		validate = append(validate, messageToImportMessage(m))
	}

	return &imports.Context{
		HasValidatableMessage:   ctx.HasValidatableMessage(),
		OutboundHasBitflagField: ctx.OutboundHasBitflagField(),
		UseCommonConverters:     ctx.UseCommonConverters(),
		ModuleName:              ctx.ModuleName,
		FullPath:                ctx.Package.FullPath,
		Methods:                 methods,
		DomainMessages:          domain,
		OutboundMessages:        outbound,
		ValidatableMessages:     validate,
		WireExtensions:          wireExtensions,
		WireInputMessages:       wireInput,
	}
}
