package mapping

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

// FieldMappingContextOptions represents common options for the different field
// mappings.
type FieldMappingContextOptions struct {
	ProtoField   *protobuf.Field   `validate:"required"`
	ProtoMessage *protobuf.Message `validate:"required"`
	Settings     *settings.Settings
	Validate     *validator.Validate
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
