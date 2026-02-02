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
	minMessageNameParts = 2
)

// Field represents a field loaded from protobuf.
type Field struct {
	optional   bool
	array      bool
	Name       string
	JSONName   string
	GoName     string
	TypeName   string
	Type       descriptor.FieldDescriptorProto_Type `validate:"-"`
	Schema     *protogen.Field                      `validate:"-"`
	Proto      *descriptor.FieldDescriptorProto     `validate:"-"`
	moduleName string
}

func parseField(proto *descriptor.FieldDescriptorProto, schema *protogen.Field, moduleName string) *Field {
	return &Field{
		optional:   proto.GetProto3Optional(),
		array:      proto.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED,
		Name:       proto.GetName(),
		JSONName:   strings.ToLower(strcase.SnakeCase(proto.GetJsonName())),
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
		f.JSONName,
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

// IsProtobufWrapper determines if the Field type is a wrapper type defined in
// the google.protobuf package.
func (f *Field) IsProtobufWrapper() bool {
	re := regexp.MustCompile(`google\.protobuf\..+Value`)
	return re.MatchString(f.TypeName)
}

// GetWrapperType returns the underlying type of protobuf wrapper if the Field
// is a protobuf wrapper; otherwise, returns an empty string.
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
		if len(parts) >= minMessageNameParts {
			module := parts[len(parts)-2]
			typeName := parts[len(parts)-1]
			return module, typeName, f.moduleName == module
		}
	}

	return "", "", false
}

// IsEnum checks if the Field is of 'enum' type.
func (f *Field) IsEnum() bool {
	return f.Type == descriptor.FieldDescriptorProto_TYPE_ENUM
}

// EnumPackage returns the enum package name, its name, and a flag indicating if
// it belongs to the current package or not.
func (f *Field) EnumPackage() (string, string, bool) {
	if f.IsEnum() {
		parts := strings.Split(f.TypeName, ".")
		if len(parts) >= minMessageNameParts {
			module := parts[len(parts)-2]
			typeName := parts[len(parts)-1]
			return module, typeName, f.moduleName == module
		}
	}

	return "", "", false
}

// IsMessage checks if the field is of the type 'message'.
func (f *Field) IsMessage() bool {
	return f.Type == descriptor.FieldDescriptorProto_TYPE_MESSAGE && !f.IsMap()
}

// IsMap determines if the field is declared as a map type in the protobuf
// schema.
func (f *Field) IsMap() bool {
	return f.Schema.Desc.IsMap()
}

// MapKeyType returns the type of the key for a map field if the field is a map;
// otherwise, it returns an empty string.
func (f *Field) MapKeyType() string {
	if !f.IsMap() {
		return ""
	}

	t := f.Schema.Desc.MapKey()
	return t.Kind().String()
}

// MapValueType returns the type of the value for a map field or an empty string
// if the field is not a map.
func (f *Field) MapValueType() string {
	if !f.IsMap() {
		return ""
	}

	t := f.Schema.Desc.MapValue()

	switch t.Kind() {
	case protoreflect.MessageKind:
		return string(t.Message().FullName())
	case protoreflect.EnumKind:
		// Enums should be mapped as string
		return "string"
	default:
		return t.Kind().String()
	}
}

// ModuleName returns the name of the module associated with the field.
func (f *Field) ModuleName() string {
	return f.moduleName
}

// MapValueTypeKind returns the protobuf kind of the value type for a map field.
func (f *Field) MapValueTypeKind() protoreflect.Kind {
	if !f.IsMap() {
		return 0
	}

	t := f.Schema.Desc.MapValue()
	return t.Kind()
}

// MapValueTypeName returns the fully qualified name of the value type if the
// field is a map; otherwise, returns an empty string.
func (f *Field) MapValueTypeName() string {
	if !f.IsMap() {
		return ""
	}

	t := f.Schema.Desc.MapValue()

	switch t.Kind() {
	case protoreflect.MessageKind:
		return string(t.Message().FullName())
	case protoreflect.EnumKind:
		return string(t.Enum().FullName())
	default:
		return t.Kind().String()
	}
}

// MapValuePackage extracts and returns the module name, type name, and a flag
// indicating if the map value is in the current package.
func (f *Field) MapValuePackage() (string, string, bool) {
	if !f.IsMap() {
		return "", "", false
	}

	// Notice that maps don't start with a leading dot  '.', that's why we're
	// not removing from here.
	parts := strings.Split(f.MapValueTypeName(), ".")
	if len(parts) != minMessageNameParts {
		return "", "", false
	}

	module := parts[len(parts)-2]
	typeName := parts[len(parts)-1]
	return module, typeName, f.moduleName == module
}
