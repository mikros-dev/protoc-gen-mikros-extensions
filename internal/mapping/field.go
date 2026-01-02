package mapping

import (
	"fmt"

	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/validation"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

// Field is the object used to make conversions between field types.
type Field struct {
	isArray       bool
	isHTTPService bool
	proto         *protobuf.Field
	validation    *validation.Call
	tag           *FieldTag
	naming        *FieldNaming
	fieldType     *FieldType
	conversion    *FieldConversion
}

// FieldOptions is the options used to create a new field.
type FieldOptions struct {
	IsHTTPService bool
	Receiver      string
	ProtoField    *protobuf.Field
	Message       *Message
	ProtoMessage  *protobuf.Message
	Settings      *settings.Settings
}

// NewField creates a new field converter.
func NewField(options FieldOptions) (*Field, error) {
	var (
		fieldExtensions = extensions.LoadFieldExtensions(options.ProtoField.Proto)
		isArray         = options.ProtoField.Proto.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
		goName          = options.ProtoField.Schema.GoName
		goType          = ProtoTypeToGoType(
			options.ProtoField.Schema.Desc.Kind(),
			options.ProtoField.Proto.GetTypeName(),
			options.ProtoMessage.ModuleName,
		)
	)

	fieldType := newFieldType(&FieldTypeOptions{
		IsArray:         isArray,
		GoType:          goType,
		Message:         options.Message,
		Protobuf:        options.ProtoField,
		FieldExtensions: fieldExtensions,
	})

	fieldNaming := newFieldNaming(&FieldNameOptions{
		GoName:            goName,
		FieldExtensions:   fieldExtensions,
		MessageExtensions: extensions.LoadMessageExtensions(options.ProtoMessage.Proto),
	})

	call, err := newValidationCall(options, isArray, fieldType, fieldExtensions)
	if err != nil {
		return nil, err
	}

	var databaseKind string
	if options.Settings != nil {
		databaseKind = options.Settings.Database.Kind
	}

	field := &Field{
		isArray:       isArray,
		isHTTPService: options.IsHTTPService,
		proto:         options.ProtoField,
		validation:    call,
		tag: newFieldTag(&FieldTagOptions{
			DatabaseKind:    databaseKind,
			FieldExtensions: fieldExtensions,
		}),
		naming:    fieldNaming,
		fieldType: fieldType,
		conversion: newFieldConversion(&FieldConversionOptions{
			MessageReceiver: options.Receiver,
			GoName:          goName,
			GoType:          goType,
			Protobuf:        options.ProtoField,
			Settings:        options.Settings,
			FieldExtensions: fieldExtensions,
			FieldNaming:     fieldNaming,
		}),
	}

	return field, nil
}

func newValidationCall(
	options FieldOptions,
	isArray bool,
	ft *FieldType,
	ext *extensions.MikrosFieldExtensions,
) (*validation.Call, error) {
	if options.Settings == nil {
		return nil, nil
	}

	return validation.NewCall(&validation.CallOptions{
		IsArray:   isArray,
		IsMessage: options.ProtoField.IsMessage(),
		ProtoName: options.ProtoField.Name,
		Receiver:  options.Receiver,
		ProtoType: ft.WireType(false),
		Options:   ext,
		Settings:  options.Settings,
		Message:   options.ProtoMessage,
	})
}

// WireType returns the current field type corresponding to the wire type.
func (f *Field) WireType(isPointer bool) string {
	return f.fieldType.WireType(isPointer)
}

// DomainType returns the current field type for the domain.
func (f *Field) DomainType(isPointer bool) string {
	return f.fieldType.DomainType(isPointer)
}

// DomainTypeForTest returns the current field type for testing templates for
// the domain.
func (f *Field) DomainTypeForTest(isPointer bool) string {
	return f.fieldType.DomainTypeForTest(isPointer)
}

// OutboundType returns the current field type for the outbound response.
func (f *Field) OutboundType(isPointer bool) string {
	return f.fieldType.OutboundType(isPointer)
}

// DomainName returns the domain name associated with the field. It is formatted
// in UpperCamelCase if annotations is available, or the Go name.
func (f *Field) DomainName() string {
	return f.naming.DomainName()
}

// DomainTag generates the struct tag string for the field based on its domain
// and naming conventions. It also adds the database struct tag if available.
func (f *Field) DomainTag() string {
	return f.tag.DomainTag(f.naming.ResolveDomainNameForTag(f.naming.DomainName()))
}

// InboundTag generates and returns the struct tag string for the inbound
// structure.
func (f *Field) InboundTag() string {
	return f.tag.InboundTag(f.naming.InboundName())
}

// InboundName returns the inbound name of the field, defaulting to snake_case
// unless overwritten by specific extensions.
func (f *Field) InboundName() string {
	return f.naming.InboundName()
}

// OutboundTag generates the outbound struct tag for the field based on its
// domain name and outbound configuration options.
func (f *Field) OutboundTag() string {
	return f.tag.OutboundTag(f.naming.ResolveOutboundNameForTag(f.naming.OutboundName()))
}

// OutboundName returns the outbound field name.
func (f *Field) OutboundName() string {
	return f.naming.OutboundName()
}

// OutboundJSONTagFieldName generates and returns the outbound JSON tag name
// for the field.
func (f *Field) OutboundJSONTagFieldName() string {
	return f.naming.OutboundJSONTagFieldName()
}

// ConvertToWireType converts the field to its wire-compatible type based on
// protobuf schema and field settings.
func (f *Field) ConvertToWireType(wireInput bool) string {
	return f.conversion.ToWireType(wireInput)
}

// ConvertDomainTypeToWireType converts the domain type into its wire-protocol
// representation.
func (f *Field) ConvertDomainTypeToWireType() string {
	return f.conversion.DomainTypeToWireType()
}

// ConvertDomainTypeToArrayWireType converts the domain-specific representation to
// its array wire format as a string.
func (f *Field) ConvertDomainTypeToArrayWireType(receiver string, wireInput bool) string {
	return f.conversion.DomainTypeToArrayWireType(receiver, wireInput)
}

// ConvertWireTypeToArrayDomainType converts a wire type into a
// domain-specific representation for array fields.
func (f *Field) ConvertWireTypeToArrayDomainType(receiver string) string {
	return f.conversion.WireTypeToArrayDomainType(receiver)
}

// ConvertDomainTypeToMapWireType converts a domain type to its corresponding
// map wire type representation.
func (f *Field) ConvertDomainTypeToMapWireType(receiver string, wireInput bool) string {
	return f.conversion.DomainTypeToMapWireType(receiver, wireInput)
}

// ConvertWireTypeToMapDomainType converts a wire type value into the corresponding
// map domain type representation.
func (f *Field) ConvertWireTypeToMapDomainType(receiver string) string {
	return f.conversion.WireTypeToMapDomainType(receiver)
}

// ConvertWireOutputToOutbound converts the field's wire format output into the
// outbound.
func (f *Field) ConvertWireOutputToOutbound(receiver string) string {
	return f.conversion.WireOutputToOutbound(receiver)
}

// ConvertWireOutputToMapOutbound converts the field wire output to its map
// outbound representation.
func (f *Field) ConvertWireOutputToMapOutbound(receiver string) string {
	return f.conversion.WireOutputToMapOutbound(receiver)
}

// ConvertWireOutputToArrayOutbound converts the field wire format output into
// an appropriate outbound array representation.
func (f *Field) ConvertWireOutputToArrayOutbound(receiver string) string {
	return f.conversion.WireOutputToArrayOutbound(receiver)
}

// ValidationName constructs and returns the validation call name for the
// field.
func (f *Field) ValidationName(receiver string) string {
	var address string
	if f.needsAddressNotation() {
		address = "&"
	}

	return fmt.Sprintf("%s%s.%s", address, receiver, f.naming.GoName)
}

func (f *Field) needsAddressNotation() bool {
	if !f.isHTTPService {
		// Non-HTTP services always need the address notation
		return true
	}

	return !f.isArray &&
		!f.proto.IsProtoStruct() &&
		!f.proto.IsProtobufWrapper() &&
		!f.proto.IsMessage() &&
		!f.proto.IsOptional()
}

// ValidationCall retrieves the validation API call from the field's validation
// if it exists.
func (f *Field) ValidationCall() string {
	if f.validation == nil {
		return ""
	}

	return f.validation.APICall()
}
