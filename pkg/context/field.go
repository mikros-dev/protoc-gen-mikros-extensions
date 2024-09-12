package context

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/converters"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/testing"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/settings"
)

type Field struct {
	IsMessage       bool
	IsMap           bool
	IsArray         bool
	IsProtoOptional bool
	Type            descriptor.FieldDescriptorProto_Type
	GoType          string
	GoName          string
	JsonName        string
	ProtoName       string
	DomainName      string
	DomainTag       string
	InboundTag      string
	OutboundName    string
	OutboundTag     string
	OutboundTagName string
	MessageReceiver string
	Location        FieldLocation
	ProtoField      *protobuf.Field

	moduleName string
	converter  *converters.Field
	testing    *testing.Field
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
		goName  = opt.Field.Schema.GoName
		goType  = converters.ProtoTypeToGoType(opt.Field.Schema.Desc.Kind(), opt.Field.Proto.GetTypeName(), opt.ModuleName)
	)

	converter, err := converters.NewField(converters.FieldOptions{
		IsArray:       isArray,
		IsHTTPService: opt.IsHTTPService,
		GoType:        goType,
		GoName:        goName,
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
		IsMessage:       opt.Field.IsMessage(),
		IsMap:           opt.Field.IsMap(),
		IsArray:         isArray,
		IsProtoOptional: opt.Field.Proto.GetProto3Optional(),
		Type:            opt.Field.Proto.GetType(),
		GoType:          goType,
		GoName:          goName,
		JsonName:        strings.ToLower(strcase.ToSnake(opt.Field.Proto.GetJsonName())),
		ProtoName:       opt.Field.Proto.GetName(),
		DomainName:      converter.DomainName(),
		DomainTag:       converter.DomainTag(),
		InboundTag:      converter.InboundTag(),
		OutboundName:    converter.OutboundName(),
		OutboundTag:     converter.OutboundTag(),
		OutboundTagName: converter.OutboundTagName(),
		Location:        getFieldLocation(opt.Field.Proto, opt.Endpoint),
		moduleName:      opt.ModuleName,
		MessageReceiver: opt.Receiver,
		ProtoField:      opt.Field,
		converter:       converter,
		testing: testing.NewField(&testing.NewFieldOptions{
			IsArray:        isArray,
			GoType:         goType,
			ProtoField:     opt.Field,
			Settings:       opt.Settings,
			FieldConverter: converter,
		}),
	}
	if err := field.Validate(); err != nil {
		return nil, err
	}

	return field, nil
}

func (f *Field) Validate() error {
	if f.isBitflag() && f.GoType != "uint64" {
		return fmt.Errorf("field '%s' has an unsupported type '%s' to be a bitflag", f.GoName, f.GoType)
	}

	return nil
}

func (f *Field) isBitflag() bool {
	if outbound := extensions.LoadFieldOutbound(f.ProtoField.Proto); outbound != nil {
		return outbound.Bitflag != nil
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
	return f.IsScalar() || (f.ProtoField.IsTimestamp() && !f.IsArray) || f.ProtoField.IsProtoStruct()
}

func (f *Field) DomainType() string {
	return f.converter.DomainType(f.IsPointer())
}

func (f *Field) ConvertDomainTypeToWireType() string {
	return f.converter.ConvertToWireType()
}

func (f *Field) WireType() string {
	return f.converter.WireType(f.IsPointer())
}

func (f *Field) OutboundType() string {
	return f.converter.OutboundType(f.IsPointer())
}

func (f *Field) ConvertDomainTypeToArrayWireType(receiver string) string {
	return f.converter.ConvertDomainTypeToArrayWireType(receiver)
}

func (f *Field) ConvertDomainTypeToMapWireType(receiver string) string {
	return f.converter.ConvertDomainTypeToMapWireType(receiver)
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
	if outbound := extensions.LoadFieldOutbound(f.ProtoField.Proto); outbound != nil {
		return outbound.GetHide()
	}

	return false
}

func (f *Field) IsOutboundBitflag() bool {
	if outbound := extensions.LoadFieldOutbound(f.ProtoField.Proto); outbound != nil {
		return outbound.GetBitflag() != nil
	}

	return false
}

func (f *Field) IsProtobufValue() bool {
	return f.ProtoField.IsProtoValue()
}

func (f *Field) IsValidatable() bool {
	validate := extensions.LoadFieldValidate(f.ProtoField.Proto)
	return validate != nil
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
