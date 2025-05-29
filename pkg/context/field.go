package context

import (
	"fmt"
	"strings"

	"github.com/stoewer/go-strcase"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/testing"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/translation"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/converters"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	tpl_types "github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/types"
)

type Field struct {
	IsMessage                bool
	IsMap                    bool
	IsArray                  bool
	IsProtoOptional          bool
	Type                     descriptor.FieldDescriptorProto_Type
	GoType                   string
	GoName                   string
	JsonName                 string
	ProtoName                string
	DomainName               string
	DomainTag                string
	InboundTag               string
	OutboundName             string
	OutboundTag              string
	OutboundJsonTagFieldName string
	MessageReceiver          string
	Location                 FieldLocation
	ProtoField               *protobuf.Field

	moduleName string
	converter  *converters.Field
	testing    *testing.Field
	extensions *mikros_extensions.MikrosFieldExtensions
}

type LoadFieldOptions struct {
	IsHTTPService    bool
	ModuleName       string
	Receiver         string
	Field            *protobuf.Field
	Message          *protobuf.Message
	Endpoint         *Endpoint
	MessageConverter *converters.Message
	Settings         *settings.Settings
}

func loadField(opt LoadFieldOptions) (*Field, error) {
	var (
		isArray = opt.Field.Proto.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
		goType  = converters.ProtoTypeToGoType(opt.Field.Schema.Desc.Kind(), opt.Field.Proto.GetTypeName(), opt.ModuleName)
	)

	converter, err := converters.NewField(converters.FieldOptions{
		IsHTTPService: opt.IsHTTPService,
		Receiver:      opt.Receiver,
		ProtoField:    opt.Field,
		Message:       opt.MessageConverter,
		ProtoMessage:  opt.Message,
		Settings:      opt.Settings,
	})
	if err != nil {
		return nil, err
	}

	field := &Field{
		IsMessage:                opt.Field.IsMessage(),
		IsMap:                    opt.Field.IsMap(),
		IsArray:                  isArray,
		IsProtoOptional:          opt.Field.Proto.GetProto3Optional(),
		Type:                     opt.Field.Proto.GetType(),
		GoType:                   goType,
		GoName:                   opt.Field.Schema.GoName,
		JsonName:                 strings.ToLower(strcase.SnakeCase(opt.Field.Proto.GetJsonName())),
		ProtoName:                opt.Field.Proto.GetName(),
		DomainName:               converter.DomainName(),
		DomainTag:                converter.DomainTag(),
		InboundTag:               converter.InboundTag(),
		OutboundName:             converter.OutboundName(),
		OutboundTag:              converter.OutboundTag(),
		OutboundJsonTagFieldName: converter.OutboundJsonTagFieldName(),
		MessageReceiver:          opt.Receiver,
		Location:                 getFieldLocation(opt.Field.Proto, opt.Endpoint),
		ProtoField:               opt.Field,
		moduleName:               opt.ModuleName,
		converter:                converter,
		testing: testing.NewField(&testing.NewFieldOptions{
			IsArray:        isArray,
			GoType:         goType,
			ProtoField:     opt.Field,
			Settings:       opt.Settings,
			FieldConverter: converter,
		}),
		extensions: mikros_extensions.LoadFieldExtensions(opt.Field.Proto),
	}
	if err := field.Validate(); err != nil {
		return nil, err
	}

	return field, nil
}

func (f *Field) Validate() error {
	if f.isBitflag() && f.GoType != "uint64" {
		return fmt.Errorf("field '%s' has an unsupported type '%s' to be a bitflag", f.ProtoName, f.GoType)
	}

	if f.hasJsonStructTag() {
		return fmt.Errorf("field '%s' cannot have a custom json struct tag", f.ProtoName)
	}

	return nil
}

func (f *Field) isBitflag() bool {
	return f.extensions != nil && f.extensions.GetOutbound() != nil && f.extensions.GetOutbound().GetBitflag() != nil
}

func (f *Field) hasJsonStructTag() bool {
	if f.extensions == nil {
		return false
	}

	if domain := f.extensions.GetDomain(); domain != nil {
		for _, st := range domain.GetStructTag() {
			if strings.Contains(st.GetName(), "json") {
				return true
			}
		}
	}

	return false
}

func (f *Field) IsPointer() bool {
	return (f.IsProtoOptional && !f.ProtoField.IsProtoStruct()) || f.ProtoField.IsProtobufWrapper() || f.IsMessage
}

func (f *Field) IsScalar() bool {
	if (f.IsMessage && !f.ProtoField.IsTimestamp()) || f.IsMap || f.IsArray {
		return false
	}

	return true
}

func (f *Field) IsBindable() bool {
	if f.hasCustomBind() {
		return false
	}

	return f.isBindableType() && !f.IsArray && !f.IsMap
}

func (f *Field) hasCustomBind() bool {
	if f.extensions == nil {
		return false
	}
	outbound := f.extensions.GetOutbound()
	return outbound != nil && outbound.GetCustomBind()
}

func (f *Field) isBindableType() bool {
	return f.IsScalar() ||
		f.ProtoField.IsTimestamp() ||
		f.ProtoField.IsProtoStruct() ||
		f.ProtoField.IsMessageFromPackage() ||
		f.IsMessageFromOtherPackage()
}

func (f *Field) IsMessageFromOtherPackage() bool {
	otherTypes := f.ProtoField.IsTimestamp() || f.ProtoField.IsProtoStruct() || f.ProtoField.IsProtoValue()
	return f.IsMessage && !otherTypes
}

func (f *Field) DomainType() string {
	return f.converter.DomainType(f.IsPointer())
}

func (f *Field) ConvertDomainTypeToWireType() string {
	return f.converter.ConvertToWireType(false)
}

func (f *Field) ConvertDomainTypeToWireInputType() string {
	return f.converter.ConvertToWireType(true)
}

func (f *Field) WireType() string {
	return f.converter.WireType(f.IsPointer())
}

func (f *Field) ConvertWireTypeToDomainType() string {
	return f.converter.ConvertDomainTypeToWireType()
}

func (f *Field) OutboundType() string {
	return f.converter.OutboundType(f.IsPointer())
}

func (f *Field) ConvertDomainTypeToArrayWireType(receiver string) string {
	return f.converter.ConvertDomainTypeToArrayWireType(receiver, false)
}

func (f *Field) ConvertWireTypeToArrayDomainType(receiver string) string {
	return f.converter.ConvertWireTypeToArrayDomainType(receiver)
}

func (f *Field) ConvertDomainTypeToArrayWireInputType(receiver string) string {
	return f.converter.ConvertDomainTypeToArrayWireType(receiver, true)
}

func (f *Field) ConvertDomainTypeToMapWireType(receiver string) string {
	return f.converter.ConvertDomainTypeToMapWireType(receiver, false)
}

func (f *Field) ConvertDomainTypeToMapWireInputType(receiver string) string {
	return f.converter.ConvertDomainTypeToMapWireType(receiver, true)
}

func (f *Field) ConvertWireTypeToMapDomainType(receiver string) string {
	return f.converter.ConvertWireTypeToMapDomainType(receiver)
}

func (f *Field) ConvertWireOutputToOutbound(receiver string) string {
	return f.converter.ConvertWireOutputToOutbound(receiver)
}

func (f *Field) ConvertWireOutputToMapOutbound(receiver string) string {
	return f.converter.ConvertWireOutputToMapOutbound(receiver)
}

func (f *Field) ConvertWireOutputToArrayOutbound(receiver string) string {
	return f.converter.ConvertWireOutputToArrayOutbound(receiver)
}

func (f *Field) OutboundHide() bool {
	return f.extensions != nil && f.extensions.GetOutbound() != nil && f.extensions.GetOutbound().GetHide()
}

func (f *Field) IsOutboundBitflag() bool {
	return f.extensions != nil && f.extensions.GetOutbound() != nil && f.extensions.GetOutbound().GetBitflag() != nil
}

func (f *Field) IsProtobufValue() bool {
	return f.ProtoField.IsProtoValue()
}

func (f *Field) IsValidatable() bool {
	return f.extensions != nil && f.extensions.GetValidate() != nil && !f.extensions.GetValidate().GetSkip()
}

func (f *Field) ValidationName(receiver string) string {
	return f.converter.ValidationName(receiver)
}

func (f *Field) ValidationCall() string {
	return f.converter.ValidationCall()
}

func (f *Field) TestingValueBinding() string {
	return f.testing.BindingValue(f.IsPointer())
}

func (f *Field) TestingValueCall() string {
	return f.testing.ValueInitCall(f.IsPointer())
}

func (f *Field) TypeByTemplateKind(kind tpl_types.Kind) string {
	if kind == tpl_types.KindRust {
		return translation.RustFieldType(
			f.Type,
			f.IsProtoOptional,
			f.IsArray,
			f.converter.WireType(false),
			f.ProtoField,
		)
	}

	return f.GoType
}

func (f *Field) HeaderArgumentByTemplateKind(kind tpl_types.Kind) string {
	if kind == tpl_types.KindRust {
		return translation.RustHeaderArgument(f.Type, f.ProtoName)
	}

	return ""
}

func (f *Field) ConvertWireOutputToOutboundByTemplateKind(kind tpl_types.Kind, receiver string) string {
	if kind == tpl_types.KindRust {
		return translation.RustWireOutputToOutbound(f.Type, f.ProtoName, receiver)
	}

	return ""
}
