package mapping

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
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
		fieldExtensions = loadFieldExtensions(options.ProtoField)
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
		MessageExtensions: loadMessageExtensions(options.ProtoMessage),
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
			DatabaseKind:      databaseKind,
			DomainName:        fieldNaming.Domain(),
			OutboundName:      fieldNaming.Outbound(),
			InboundName:       fieldNaming.Inbound(),
			FieldExtensions:   fieldExtensions,
			MessageExtensions: loadMessageExtensions(options.ProtoMessage),
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

func loadFieldExtensions(proto *protobuf.Field) *extensions.MikrosFieldExtensions {
	ext := extensions.LoadFieldExtensions(proto.Proto)
	if ext == nil {
		// We return an empty struct here so we don't need to always check for
		// nil. But its sub-messages will be nil as well, so they must be
		// validated.
		return &extensions.MikrosFieldExtensions{}
	}

	return ext
}

func loadMessageExtensions(proto *protobuf.Message) *extensions.MikrosMessageExtensions {
	ext := extensions.LoadMessageExtensions(proto.Proto)
	if ext == nil {
		// We return an empty struct here so we don't need to always check for
		// nil. But its sub-messages will be nil as well, so they must be
		// validated.
		return &extensions.MikrosMessageExtensions{}
	}

	return ext
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
		WireType:  ft.Wire(false),
		Options:   ext,
		Settings:  options.Settings,
		Message:   options.ProtoMessage,
	})
}

// Types returns the field type converter.
func (f *Field) Types() *FieldType {
	return f.fieldType
}

// Tags returns the field tag converter.
func (f *Field) Tags() *FieldTag {
	return f.tag
}

// Naming returns the field naming converter.
func (f *Field) Naming() *FieldNaming {
	return f.naming
}

// Conversion returns the field conversion converter.
func (f *Field) Conversion() *FieldConversion {
	return f.conversion
}

// ValidationName constructs and returns the validation call name for the
// field.
func (f *Field) ValidationName(receiver string) string {
	var address string
	if f.needsAddressNotation() {
		address = "&"
	}

	return fmt.Sprintf("%s%s.%s", address, receiver, f.naming.GoName())
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

func getMapKeyValueTypesForWire(field *protobuf.Field) (string, string, protoreflect.FieldDescriptor) {
	var (
		v     = field.Schema.Desc.MapValue()
		value = ProtoTypeToGoType(v.Kind(), "", "")
	)

	if v.Kind() == protoreflect.MessageKind {
		name := string(v.Message().Name())
		if name == "Timestamp" {
			name = "ts.Timestamp"
		}

		parts := strings.Split(string(v.Message().FullName()), ".")
		value = "*" + name
		if parts[1] != field.ModuleName() {
			value = fmt.Sprintf("*%s.%s", parts[1], v.Message().Name())
		}
	}

	if v.Kind() == protoreflect.EnumKind {
		parts := strings.Split(string(v.Enum().FullName()), ".")
		value = parts[len(parts)-1]
		if parts[1] != field.ModuleName() {
			value = fmt.Sprintf("%s.%s", parts[1], v.Enum().Name())
		}
	}

	return ProtoKindToGoType(field.Schema.Desc.MapKey().Kind()), value, v
}

func handleOtherModuleField(fieldType string, field *protobuf.Field) (string, string, bool) {
	if hasModuleAsPrefix(fieldType, field) {
		parts := strings.Split(fieldType, ".")
		if len(parts) < 2 {
			// Something is wrong here
			return "", "", false
		}

		return parts[len(parts)-2], parts[len(parts)-1], true
	}

	return "", "", false
}

func hasModuleAsPrefix(fieldType string, field *protobuf.Field) bool {
	return strings.Contains(fieldType, ".") &&
		!field.IsProtoStruct() &&
		!field.IsTimestamp() &&
		!field.IsProtoValue() &&
		!field.IsProtobufWrapper()
}
