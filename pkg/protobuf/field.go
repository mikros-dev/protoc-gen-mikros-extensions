package protobuf

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

const (
	internalMessageTypeParts = 4
)

type Field struct {
	optional   bool
	array      bool
	Name       string
	JsonName   string
	GoName     string
	TypeName   string
	Type       descriptor.FieldDescriptorProto_Type
	Schema     *protogen.Field
	Proto      *descriptor.FieldDescriptorProto
	moduleName string
}

func parseField(proto *descriptor.FieldDescriptorProto, schema *protogen.Field, moduleName string) *Field {
	return &Field{
		optional:   proto.GetProto3Optional(),
		array:      proto.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED,
		Name:       proto.GetName(),
		JsonName:   strings.ToLower(strcase.SnakeCase(proto.GetJsonName())),
		GoName:     schema.GoName,
		TypeName:   proto.GetTypeName(),
		Type:       proto.GetType(),
		Schema:     schema,
		Proto:      proto,
		moduleName: moduleName,
	}
}

func (f *Field) String() string {
	return fmt.Sprintf(`{name:%v, go_name:%v, json_name:%v, type:%v, type_name:%v, optional:%v, array:%v}`,
		f.Name,
		f.GoName,
		f.JsonName,
		f.Type,
		f.TypeName,
		f.IsOptional(),
		f.IsArray())
}

// IsOptional indicates if the Field is declared as optional or not.
func (f *Field) IsOptional() bool {
	return f.optional
}

// IsArray indicates if the Field is declared as repeated or not.
func (f *Field) IsArray() bool {
	return f.array && !f.IsMap()
}

// IsTimestamp checks if the Field is of 'google.protobuf.Timestamp' type.
func (f *Field) IsTimestamp() bool {
	return f.IsMessageTypeOf(".google.protobuf.Timestamp")
}

// IsProtoStruct checks if the Field is of 'google.protobuf.Struct' type.
func (f *Field) IsProtoStruct() bool {
	return f.IsMessageTypeOf(".google.protobuf.Struct")
}

// IsProtoValue checks if the Field is of 'google.protobuf.Value' type.
func (f *Field) IsProtoValue() bool {
	return f.IsMessageTypeOf(".google.protobuf.Value")
}

// IsProtoAny checks if the Field is of 'google.protobuf.Any' type.
func (f *Field) IsProtoAny() bool {
	return f.IsMessageTypeOf(".google.protobuf.Any")
}

// IsMessageTypeOf checks if the Field is of a specific message type.
func (f *Field) IsMessageTypeOf(typeOf string) bool {
	return f.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE && f.TypeName == typeOf
}

// IsMessageFromPackage checks if the Field is a message, and it belongs to the
// current package or not.
func (f *Field) IsMessageFromPackage() bool {
	_, _, ok := f.MessagePackage()
	return ok
}

func (f *Field) IsProtobufWrapper() bool {
	re := regexp.MustCompile(`google\.protobuf\..+Value`)
	return re.MatchString(f.TypeName)
}

func (f *Field) GetWrapperType() string {
	if f.IsProtobufWrapper() {
		re := regexp.MustCompile(`google\.protobuf\.(.+)Value`)
		s := re.FindStringSubmatch(f.TypeName)
		return s[1]
	}

	return ""
}

// MessagePackage returns the message package name and a flag indicating if it
// belongs to the current package or not.
func (f *Field) MessagePackage() (string, string, bool) {
	if f.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		parts := strings.Split(f.TypeName, ".")
		if len(parts) == internalMessageTypeParts {
			module := parts[len(parts)-2]
			typeName := parts[len(parts)-1]
			return module, typeName, f.moduleName == module
		}
	}

	return "", "", false
}

func (f *Field) IsEnum() bool {
	return f.Type == descriptor.FieldDescriptorProto_TYPE_ENUM
}

func (f *Field) EnumPackage() (string, string, bool) {
	if f.IsEnum() {
		parts := strings.Split(f.TypeName, ".")
		if len(parts) == internalMessageTypeParts {
			module := parts[len(parts)-2]
			typeName := parts[len(parts)-1]
			return module, typeName, f.moduleName == module
		}
	}

	return "", "", false
}

func (f *Field) IsMessage() bool {
	return f.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE && !f.IsMap()
}

func (f *Field) IsMap() bool {
	return f.Schema.Desc.IsMap()
}

func (f *Field) MapKeyType() string {
	if f.IsMap() {
		t := f.Schema.Desc.MapKey()
		return t.Kind().String()
	}

	return ""
}

func (f *Field) MapValueType() string {
	if f.IsMap() {
		t := f.Schema.Desc.MapValue()

		if t.Kind() == protoreflect.MessageKind {
			return string(t.Message().FullName())
		}

		// Enum maps will always be a string map inside a Model structure
		if t.Kind() == protoreflect.EnumKind {
			return "string"
		}

		return t.Kind().String()
	}

	return ""
}

func (f *Field) ModuleName() string {
	return f.moduleName
}

func (f *Field) MapValueTypeKind() protoreflect.Kind {
	if f.IsMap() {
		t := f.Schema.Desc.MapValue()
		return t.Kind()
	}

	return 0
}

func (f *Field) MapValueTypeName() string {
	if f.IsMap() {
		t := f.Schema.Desc.MapValue()

		if t.Kind() == protoreflect.MessageKind {
			return string(t.Message().FullName())
		}

		if t.Kind() == protoreflect.EnumKind {
			return string(t.Enum().FullName())
		}
	}

	return ""
}

func (f *Field) MapValuePackage() (string, string, bool) {
	if f.IsMap() {
		parts := strings.Split(f.MapValueTypeName(), ".")

		// Map types don't come with the leading dot '.', that's why we're
		// subtracting one here.
		if len(parts) == internalMessageTypeParts-1 {
			module := parts[len(parts)-2]
			typeName := parts[len(parts)-1]
			return module, typeName, f.moduleName == module
		}
	}

	return "", "", false
}
