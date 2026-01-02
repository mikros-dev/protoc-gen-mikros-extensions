package mapping

import (
	"fmt"
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// TODO: Remove 'Type' from function names

type conversionMode int

const (
	wireToDomain conversionMode = iota
	wireToOutbound
)

type FieldTypeOptions struct {
	IsArray         bool
	GoType          string
	Message         *Message
	Protobuf        *protobuf.Field
	FieldExtensions *extensions.MikrosFieldExtensions
}

// FieldType is the mechanism that allows getting the Field type for
// different scenarios.
type FieldType struct {
	isArray    bool
	goType     string
	msg        *Message
	proto      *protobuf.Field
	extensions *extensions.MikrosFieldExtensions
}

func newFieldType(options *FieldTypeOptions) *FieldType {
	return &FieldType{
		isArray:    options.IsArray,
		goType:     options.GoType,
		msg:        options.Message,
		proto:      options.Protobuf,
		extensions: options.FieldExtensions,
	}
}

// WireType returns the wire type for the field.
func (f *FieldType) WireType(isPointer bool) string {
	if f.proto.IsMap() {
		key, value, _ := f.getMapKeyValueTypesForWire()
		return fmt.Sprintf("map[%s]%s", key, value)
	}

	if f.proto.IsTimestamp() {
		return formatType("ts.Timestamp", f.isArray, isPointer)
	}

	// Handle fields from other modules
	if module, name, ok := f.handleOtherModuleField(f.goType); ok {
		prefix := ""
		if module != f.proto.ModuleName() {
			prefix = fmt.Sprintf("%s.", module)
		}

		t := fmt.Sprintf("%s%s", prefix, name)
		return formatType(t, f.isArray, isPointer)
	}

	// f.goType is always the field wire type. Here we just adjust it in case
	// we're dealing with an array or pointers.
	return formatType(f.goType, f.isArray, isPointer)
}

func (f *FieldType) getMapKeyValueTypesForWire() (string, string, protoreflect.FieldDescriptor) {
	var (
		v     = f.proto.Schema.Desc.MapValue()
		value = ProtoTypeToGoType(v.Kind(), "", "")
	)

	if v.Kind() == protoreflect.MessageKind {
		name := string(v.Message().Name())
		if name == "Timestamp" {
			name = "ts.Timestamp"
		}

		parts := strings.Split(string(v.Message().FullName()), ".")
		value = "*" + name
		if parts[1] != f.proto.ModuleName() {
			value = fmt.Sprintf("*%s.%s", parts[1], v.Message().Name())
		}
	}

	if v.Kind() == protoreflect.EnumKind {
		parts := strings.Split(string(v.Enum().FullName()), ".")
		value = parts[len(parts)-1]
		if parts[1] != f.proto.ModuleName() {
			value = fmt.Sprintf("%s.%s", parts[1], v.Enum().Name())
		}
	}

	return ProtoKindToGoType(f.proto.Schema.Desc.MapKey().Kind()), value, v
}

func (f *FieldType) handleOtherModuleField(fieldType string) (string, string, bool) {
	if f.hasModuleAsPrefix(fieldType) {
		parts := strings.Split(fieldType, ".")
		if len(parts) < 2 {
			// Something is wrong here
			return "", "", false
		}

		return parts[len(parts)-2], parts[len(parts)-1], true
	}

	return "", "", false
}

func (f *FieldType) hasModuleAsPrefix(fieldType string) bool {
	return strings.Contains(fieldType, ".") &&
		!f.proto.IsProtoStruct() &&
		!f.proto.IsTimestamp() &&
		!f.proto.IsProtoValue() &&
		!f.proto.IsProtobufWrapper()
}

// DomainType returns the domain type for the field.
func (f *FieldType) DomainType(isPointer bool) string {
	return f.convertFromWireType(isPointer, false, wireToDomain)
}

// DomainTypeForTest returns the domain type for the field for testing purposes.
func (f *FieldType) DomainTypeForTest(isPointer bool) string {
	return f.convertFromWireType(isPointer, true, wireToDomain)
}

// OutboundType returns the outbound type for the field.
func (f *FieldType) OutboundType(isPointer bool) string {
	return f.convertFromWireType(isPointer, false, wireToOutbound)
}

func (f *FieldType) convertFromWireType(isPointer, testMode bool, mode conversionMode) string {
	if f.extensions != nil && mode == wireToOutbound {
		if t := f.extensions.GetOutbound().GetCustomType(); t != "" {
			return t
		}
	}

	// Handle Built-in Proto Types
	if t, ok := f.getBuiltInType(isPointer); ok {
		return t
	}

	// Handle Complex Types
	if f.proto.IsMap() {
		key, value := f.getMapKeyValueTypes(testMode, mode)
		return fmt.Sprintf("map[%s]%s", key, value)
	}

	// Handle External/Cross-Module Types
	if t, ok := f.getExternalModuleType(isPointer, testMode, mode); ok {
		return t
	}

	// Default
	baseType := f.goType
	if mode == wireToOutbound {
		baseType = f.convertFromWireTypeToOutbound()
	}

	return formatType(baseType, f.isArray, isPointer)
}

func (f *FieldType) getBuiltInType(isPointer bool) (string, bool) {
	switch {
	case f.proto.IsEnum():
		return formatType("string", f.isArray, isPointer), true
	case f.proto.IsProtoStruct():
		return formatType("map[string]interface{}", f.isArray, false), true
	case f.proto.IsTimestamp():
		return formatType("time.Time", f.isArray, isPointer), true
	case f.proto.IsProtoValue():
		return "interface{}", true
	}

	return "", false
}

func (f *FieldType) getMapKeyValueTypes(testMode bool, mode conversionMode) (string, string) {
	var (
		v     = f.proto.Schema.Desc.MapValue()
		value = ProtoTypeToGoType(v.Kind(), "", "")
	)

	if v.Kind() == protoreflect.MessageKind {
		valueType := f.msg.WireToDomainMapValueType(string(v.Message().Name()))
		if mode == wireToOutbound {
			valueType = f.msg.WireOutputToOutbound(string(v.Message().Name()))
		}

		module, _, ok := f.handleOtherModuleField(string(v.Message().FullName()))
		if ok && (module != f.proto.ModuleName() || testMode) {
			valueType = fmt.Sprintf("%s.%s", module, valueType)
		}

		value = "*" + valueType
	}

	if v.Kind() == protoreflect.EnumKind {
		value = "string"
	}

	return ProtoKindToGoType(f.proto.Schema.Desc.MapKey().Kind()), value
}

func (f *FieldType) getExternalModuleType(isPointer, testMode bool, mode conversionMode) (string, bool) {
	module, name, ok := f.handleOtherModuleField(f.goType)
	if !ok {
		return "", false
	}

	prefix := ""
	if module != f.proto.ModuleName() || testMode {
		prefix = fmt.Sprintf("%s.", module)
	}

	suffix := f.msg.WireToDomain(name)
	if mode == wireToOutbound {
		suffix = f.msg.WireOutputToOutbound(name)
	}

	return formatType(prefix+suffix, f.isArray, isPointer), true
}

func formatType(outType string, isArray, isPointer bool) string {
	ptr := ""
	if isPointer && !strings.HasPrefix(outType, "*") {
		ptr = "*"
	}

	if isArray {
		return "[]" + ptr + outType
	}

	return ptr + outType
}

func (f *FieldType) convertFromWireTypeToOutbound() string {
	if f.extensions != nil {
		if outbound := f.extensions.GetOutbound(); outbound != nil {
			// bitflag fields usually are declared as an unsigned integer, allowing
			// each one of its bits to be set to different information. Thus, the
			// outbound type should be a slice of strings like an enum list.
			if outbound.GetBitflag() != nil {
				return "[]string"
			}
		}
	}

	return f.goType
}
