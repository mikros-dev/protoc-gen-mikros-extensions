package mapping

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

// Field is an object that gathers all supported field mapping operations in
// a single one.
type Field struct {
	isArray       bool
	isHTTPService bool
	proto         *protobuf.Field
	validation    *FieldValidation
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
		messageExtensions = loadMessageExtensions(options.ProtoMessage)
		fieldExtensions   = loadFieldExtensions(options.ProtoField)
		fieldType         = NewFieldType(&FieldTypeOptions{
			Message:         options.Message,
			ProtoField:      options.ProtoField,
			ProtoMessage:    options.ProtoMessage,
			FieldExtensions: fieldExtensions,
		})

		fieldNaming = NewFieldNaming(&FieldNameOptions{
			ProtoField:        options.ProtoField,
			FieldExtensions:   fieldExtensions,
			MessageExtensions: messageExtensions,
		})
	)

	v, err := NewFieldValidation(FieldValidationOptions{
		IsHTTPService: options.IsHTTPService,
		Receiver:      options.Receiver,
		FieldNaming:   fieldNaming,
		FieldType:     fieldType,
		ProtoField:    options.ProtoField,
		ProtoMessage:  options.ProtoMessage,
		Settings:      options.Settings,
	})
	if err != nil {
		return nil, err
	}

	var databaseKind string
	if options.Settings != nil {
		databaseKind = options.Settings.Database.Kind
	}

	field := &Field{
		isArray:       options.ProtoField.IsArray(),
		isHTTPService: options.IsHTTPService,
		proto:         options.ProtoField,
		validation:    v,
		tag: NewFieldTag(&FieldTagOptions{
			DatabaseKind:      databaseKind,
			FieldExtensions:   fieldExtensions,
			FieldNaming:       fieldNaming,
			MessageExtensions: messageExtensions,
		}),
		naming:    fieldNaming,
		fieldType: fieldType,
		conversion: NewFieldConversion(&FieldConversionOptions{
			MessageReceiver: options.Receiver,
			Protobuf:        options.ProtoField,
			Settings:        options.Settings,
			FieldExtensions: fieldExtensions,
			FieldNaming:     fieldNaming,
			FieldType:       fieldType,
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

// Validation returns the field validation converter.
func (f *Field) Validation() *FieldValidation {
	return f.validation
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
