package context

import (
	"fmt"
	"strings"

	"github.com/stoewer/go-strcase"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/testing"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mapping"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

// Field represents a field to be used inside templates by its context.
type Field struct {
	IsMessage                bool
	IsMap                    bool
	IsArray                  bool
	IsProtoOptional          bool
	Type                     descriptor.FieldDescriptorProto_Type
	GoType                   string
	GoName                   string
	JSONName                 string
	ProtoName                string
	DomainName               string
	DomainTag                string
	InboundTag               string
	OutboundName             string
	OutboundTag              string
	OutboundJSONTagFieldName string
	MessageReceiver          string
	Location                 FieldLocation
	ProtoField               *protobuf.Field

	moduleName string
	converter  *mapping.Field
	testing    *testing.Field
	extensions *extensions.MikrosFieldExtensions
}

type loadFieldOptions struct {
	IsHTTPService    bool
	ModuleName       string
	Receiver         string
	Field            *protobuf.Field
	Message          *protobuf.Message
	Endpoint         *Endpoint
	MessageConverter *mapping.Message
	Settings         *settings.Settings
}

func loadField(opt loadFieldOptions) (*Field, error) {
	var (
		isArray = opt.Field.Proto.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
		goType  = mapping.ProtoTypeToGoType(
			opt.Field.Schema.Desc.Kind(),
			opt.Field.Proto.GetTypeName(),
			opt.ModuleName,
		)
	)

	converter, err := mapping.NewField(mapping.FieldOptions{
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
		JSONName:                 strings.ToLower(strcase.SnakeCase(opt.Field.Proto.GetJsonName())),
		ProtoName:                opt.Field.Proto.GetName(),
		DomainName:               converter.DomainName(),
		DomainTag:                converter.DomainTag(),
		InboundTag:               converter.InboundTag(),
		OutboundName:             converter.OutboundName(),
		OutboundTag:              converter.OutboundTag(),
		OutboundJSONTagFieldName: converter.OutboundJSONTagFieldName(),
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
		extensions: extensions.LoadFieldExtensions(opt.Field.Proto),
	}
	if err := field.Validate(); err != nil {
		return nil, err
	}

	return field, nil
}

// Validate checks if the field is valid.
func (f *Field) Validate() error {
	if f.isBitflag() && f.GoType != "uint64" {
		return fmt.Errorf("field '%s' has an unsupported type '%s' to be a bitflag", f.ProtoName, f.GoType)
	}

	if f.hasJSONStructTag() {
		return fmt.Errorf("field '%s' cannot have a custom json struct tag", f.ProtoName)
	}

	return nil
}

func (f *Field) isBitflag() bool {
	return f.extensions != nil && f.extensions.GetOutbound() != nil && f.extensions.GetOutbound().GetBitflag() != nil
}

func (f *Field) hasJSONStructTag() bool {
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

// IsPointer returns true if the field is a pointer.
func (f *Field) IsPointer() bool {
	return (f.IsProtoOptional && !f.ProtoField.IsProtoStruct()) || f.ProtoField.IsProtobufWrapper() || f.IsMessage
}

// IsScalar returns true if the field is a scalar.
func (f *Field) IsScalar() bool {
	if (f.IsMessage && !f.ProtoField.IsTimestamp()) || f.IsMap || f.IsArray {
		return false
	}

	return true
}

// IsBindable returns true if the field is bindable, i.e., it can be used in
// instructions where a value is bound into it.
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

// IsMessageFromOtherPackage returns true if the field is a message from another
// package.
func (f *Field) IsMessageFromOtherPackage() bool {
	otherTypes := f.ProtoField.IsTimestamp() || f.ProtoField.IsProtoStruct() || f.ProtoField.IsProtoValue()
	return f.IsMessage && !otherTypes
}

// DomainType returns the domain type of the field.
func (f *Field) DomainType() string {
	return f.converter.DomainType(f.IsPointer())
}

// ConvertDomainTypeToWireType converts the domain type to the wire type.
func (f *Field) ConvertDomainTypeToWireType() string {
	return f.converter.ConvertToWireType(false)
}

// ConvertDomainTypeToWireInputType converts the domain type to the wire input
// type.
func (f *Field) ConvertDomainTypeToWireInputType() string {
	return f.converter.ConvertToWireType(true)
}

// WireType returns the wire type of the field.
func (f *Field) WireType() string {
	return f.converter.WireType(f.IsPointer())
}

// ConvertWireTypeToDomainType converts the wire type to the domain type.
func (f *Field) ConvertWireTypeToDomainType() string {
	return f.converter.ConvertDomainTypeToWireType()
}

// OutboundType returns the outbound type of the field.
func (f *Field) OutboundType() string {
	return f.converter.OutboundType(f.IsPointer())
}

// ConvertDomainTypeToArrayWireType converts a domain type to its corresponding
// wire type.
func (f *Field) ConvertDomainTypeToArrayWireType(receiver string) string {
	return f.converter.ConvertDomainTypeToArrayWireType(receiver, false)
}

// ConvertWireTypeToArrayDomainType converts a wire type to its corresponding
// domain type.
func (f *Field) ConvertWireTypeToArrayDomainType(receiver string) string {
	return f.converter.ConvertWireTypeToArrayDomainType(receiver)
}

// ConvertDomainTypeToArrayWireInputType converts the domain type to an array
// wire input type.
func (f *Field) ConvertDomainTypeToArrayWireInputType(receiver string) string {
	return f.converter.ConvertDomainTypeToArrayWireType(receiver, true)
}

// ConvertDomainTypeToMapWireType converts a domain type to its corresponding
// map wire type representation.
func (f *Field) ConvertDomainTypeToMapWireType(receiver string) string {
	return f.converter.ConvertDomainTypeToMapWireType(receiver, false)
}

// ConvertDomainTypeToMapWireInputType converts the domain type into its
// corresponding map wire input type.
func (f *Field) ConvertDomainTypeToMapWireInputType(receiver string) string {
	return f.converter.ConvertDomainTypeToMapWireType(receiver, true)
}

// ConvertWireTypeToMapDomainType converts a wire type to its corresponding map
// domain type representation.
func (f *Field) ConvertWireTypeToMapDomainType(receiver string) string {
	return f.converter.ConvertWireTypeToMapDomainType(receiver)
}

// ConvertWireOutputToOutbound converts a wire output to its corresponding
// outbound representation.
func (f *Field) ConvertWireOutputToOutbound(receiver string) string {
	return f.converter.ConvertWireOutputToOutbound(receiver)
}

// ConvertWireOutputToMapOutbound converts the wire output to its corresponding
// map outbound representation.
func (f *Field) ConvertWireOutputToMapOutbound(receiver string) string {
	return f.converter.ConvertWireOutputToMapOutbound(receiver)
}

// ConvertWireOutputToArrayOutbound convertes the wire output representation of
// a field into its array outbound form.
func (f *Field) ConvertWireOutputToArrayOutbound(receiver string) string {
	return f.converter.ConvertWireOutputToArrayOutbound(receiver)
}

// OutboundHide returns true if the field should be hidden from the outbound
// representation.
func (f *Field) OutboundHide() bool {
	return f.extensions != nil && f.extensions.GetOutbound() != nil && f.extensions.GetOutbound().GetHide()
}

// IsOutboundBitflag returns true if the field is a bitflag field.
func (f *Field) IsOutboundBitflag() bool {
	return f.extensions != nil && f.extensions.GetOutbound() != nil && f.extensions.GetOutbound().GetBitflag() != nil
}

// IsProtobufValue returns true if the field is a protobuf value.
func (f *Field) IsProtobufValue() bool {
	return f.ProtoField.IsProtoValue()
}

// IsValidatable returns true if the field is validatable.
func (f *Field) IsValidatable() bool {
	return f.extensions != nil && f.extensions.GetValidate() != nil && !f.extensions.GetValidate().GetSkip()
}

// ValidationName returns the validation call name for the field.
func (f *Field) ValidationName(receiver string) string {
	return f.converter.ValidationName(receiver)
}

// ValidationCall returns the validation call for the field, name and arguments.
func (f *Field) ValidationCall() string {
	return f.converter.ValidationCall()
}

// TestingValueBinding returns the binding value for the field for the testing
// templates.
func (f *Field) TestingValueBinding() string {
	return f.testing.BindingValue(f.IsPointer())
}

// TestingValueCall returns the call to initialize the field for the testing
// templates.
func (f *Field) TestingValueCall() string {
	return f.testing.ValueInitCall(f.IsPointer())
}
