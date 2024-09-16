package converters

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/validation"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/settings"
)

type conversionMode int

const (
	wireToDomain conversionMode = iota
	wireToOutbound
)

type Field struct {
	isArray         bool
	isHTTPService   bool
	goType          string
	goName          string
	receiver        string
	msg             *Message
	db              *Database
	domain          *extensions.FieldDomainOptions
	inbound         *extensions.FieldInboundOptions
	outbound        *extensions.FieldOutboundOptions
	messageInbound  *extensions.MessageInboundOptions
	messageOutbound *extensions.MessageOutboundOptions
	proto           *protobuf.Field
	settings        *settings.Settings
	validation      *validation.Call
}

type FieldOptions struct {
	IsArray       bool
	IsHTTPService bool
	GoType        string
	GoName        string
	Receiver      string
	ProtoField    *protobuf.Field
	Message       *Message
	ProtoMessage  *protobuf.Message
	Settings      *settings.Settings
}

func NewField(options FieldOptions) (*Field, error) {
	call, err := validation.NewCall(&validation.CallOptions{
		IsArray:   options.IsArray,
		ProtoName: options.ProtoField.Name,
		Receiver:  options.Receiver,
		Options:   extensions.LoadFieldValidate(options.ProtoField.Proto),
		Settings:  options.Settings,
		Message:   options.ProtoMessage,
	})
	if err != nil {
		return nil, err
	}

	return &Field{
		isArray:         options.IsArray,
		isHTTPService:   options.IsHTTPService,
		goType:          options.GoType,
		goName:          options.GoName,
		receiver:        options.Receiver,
		msg:             options.Message,
		db:              databaseFromString(options.Settings.Database.Kind, extensions.LoadFieldDatabase(options.ProtoField.Proto)),
		domain:          extensions.LoadFieldDomain(options.ProtoField.Proto),
		inbound:         extensions.LoadFieldInbound(options.ProtoField.Proto),
		outbound:        extensions.LoadFieldOutbound(options.ProtoField.Proto),
		messageInbound:  extensions.LoadMessageInboundOptions(options.ProtoMessage.Proto),
		messageOutbound: extensions.LoadMessageOutboundOptions(options.ProtoMessage.Proto),
		proto:           options.ProtoField,
		settings:        options.Settings,
		validation:      call,
	}, nil
}

// WireType returns the current field type corresponding to the wire type.
func (f *Field) WireType(isPointer bool) string {
	if f.proto.IsMap() {
		key, value, _ := f.getMapKeyValueTypesForWire()
		return fmt.Sprintf("map[%s]%s", key, value)
	}

	if f.proto.IsTimestamp() {
		return optional("ts.Timestamp", f.isArray, isPointer)
	}

	// Handle fields from other modules
	if module, name, ok := f.handleOtherModuleField(f.goType); ok {
		prefix := ""
		if module != f.proto.ModuleName() {
			prefix = fmt.Sprintf("%s.", module)
		}

		t := fmt.Sprintf("%s%s", prefix, name)
		return optional(t, f.isArray, isPointer)
	}

	return optional(f.goType, f.isArray, isPointer)
}

func (f *Field) getMapKeyValueTypesForWire() (string, string, protoreflect.FieldDescriptor) {
	var (
		key   = f.proto.Schema.Desc.MapKey().Kind().String()
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

	return key, value, v
}

// DomainType returns the current field type for the domain.
func (f *Field) DomainType(isPointer bool) string {
	return f.convertFromWireType(isPointer, false, wireToDomain)
}

func (f *Field) DomainTypeForTest(isPointer bool) string {
	return f.convertFromWireType(isPointer, true, wireToDomain)
}

// OutboundType returns the current field type for the outbound response.
func (f *Field) OutboundType(isPointer bool) string {
	return f.convertFromWireType(isPointer, false, wireToOutbound)
}

func (f *Field) convertFromWireType(isPointer, testMode bool, mode conversionMode) string {
	if f.proto.IsEnum() {
		return optional("string", f.isArray, isPointer)
	}

	if f.proto.IsProtoStruct() {
		return "map[string]interface{}"
	}

	if f.proto.IsTimestamp() {
		return optional("time.Time", f.isArray, isPointer)
	}

	if f.proto.IsProtoValue() {
		return "interface{}"
	}

	if f.proto.IsMap() {
		key, value := f.getMapKeyValueTypes(testMode, mode)
		return fmt.Sprintf("map[%s]%s", key, value)
	}

	// Handle fields from other modules
	if module, name, ok := f.handleOtherModuleField(f.goType); ok {
		prefix := ""
		if module != f.proto.ModuleName() || testMode {
			prefix = fmt.Sprintf("%s.", module)
		}

		suffix := f.msg.WireToDomain(name)
		if mode == wireToOutbound {
			suffix = f.msg.WireOutputToOutbound(name)
		}

		t := fmt.Sprintf("%s%s", prefix, suffix)
		return optional(t, f.isArray, isPointer)
	}

	// Handle outbound specific types
	if mode == wireToOutbound {
		return optional(f.convertFromWireTypeToOutbound(), f.isArray, isPointer)
	}

	return optional(f.goType, f.isArray, isPointer)
}

func (f *Field) convertFromWireTypeToOutbound() string {
	if f.outbound != nil {
		if f.outbound.GetBitflag() != nil {
			return "[]string"
		}
	}

	return f.goType
}

func (f *Field) getMapKeyValueTypes(testMode bool, mode conversionMode) (string, string) {
	var (
		key   = f.proto.Schema.Desc.MapKey().Kind().String()
		v     = f.proto.Schema.Desc.MapValue()
		value = ProtoTypeToGoType(v.Kind(), "", "")
	)

	if v.Kind() == protoreflect.MessageKind {
		valueType := f.msg.WireToDomainMapValueType(string(v.Message().Name()))
		if mode == wireToOutbound {
			valueType = f.msg.WireOutputToOutbound(string(v.Message().Name()))
		}

		if module, _, ok := f.handleOtherModuleField(string(v.Message().FullName())); ok && (module != f.proto.ModuleName() || testMode) {
			valueType = fmt.Sprintf("%s.%s", module, valueType)
		}

		value = "*" + valueType
	}

	if v.Kind() == protoreflect.EnumKind {
		value = "string"
	}

	return key, value
}

func array(outType string, isArray bool) string {
	if isArray {
		return "[]" + outType
	}

	return outType
}

func optional(outType string, isArray, isPointer bool) string {
	if isPointer {
		return array("*"+outType, isArray)
	}

	return array(outType, isArray)
}

func (f *Field) handleOtherModuleField(fieldType string) (string, string, bool) {
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

func (f *Field) hasModuleAsPrefix(fieldType string) bool {
	return strings.Contains(fieldType, ".") &&
		!f.proto.IsProtoStruct() &&
		!f.proto.IsTimestamp() &&
		!f.proto.IsProtoValue() &&
		!f.proto.IsProtobufWrapper()
}

func (f *Field) DomainName() string {
	if f.domain != nil {
		if n := f.domain.GetName(); n != "" {
			return strcase.ToCamel(n)
		}
	}

	return f.goName
}

func (f *Field) OutboundName() string {
	return f.goName
}

func (f *Field) DomainTag() string {
	fieldName := strcase.ToSnake(f.DomainName())
	return fmt.Sprintf("`json:\"%s,omitempty\" %s`", fieldName, f.db.Tag(fieldName))
}

func (f *Field) InboundTag() string {
	name := f.DomainName()
	if f.inbound != nil {
		if n := f.inbound.GetName(); n != "" {
			name = n
		}
	}

	// Default is snake_case
	fieldName := strcase.ToSnake(name)
	if f.messageInbound != nil {
		if f.messageInbound.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
			fieldName = inboundOutboundCamelCase(name)
		}
	}

	return fmt.Sprintf("`json:\"%s\"`", fieldName)
}

func (f *Field) OutboundTag() string {
	var (
		name = f.DomainName()
		tag  = ",omitempty"
	)

	if f.outbound != nil {
		if n := f.outbound.GetName(); n != "" {
			name = n
		}
		if f.outbound.GetAllowEmpty() {
			tag = ""
		}
	}

	// Default is snake_case
	fieldName := strcase.ToSnake(name)
	if f.messageOutbound != nil {
		if f.messageOutbound.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
			fieldName = inboundOutboundCamelCase(name)
		}
	}

	return fmt.Sprintf("`json:\"%s%s\"`", fieldName, tag)
}

func (f *Field) OutboundTagName() string {
	var (
		name = f.DomainName()
	)

	if f.outbound != nil {
		if n := f.outbound.GetName(); n != "" {
			name = n
		}
	}

	// Default is snake_case
	fieldName := strcase.ToSnake(name)
	if f.messageOutbound != nil {
		if f.messageOutbound.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
			fieldName = inboundOutboundCamelCase(name)
		}
	}

	return fieldName
}

func (f *Field) ConvertToWireType() string {
	if f.proto.IsEnum() {
		return f.enumWireType()
	}

	if f.proto.IsProtoValue() {
		call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallToProtoValue)
		return fmt.Sprintf("%s(%s.%s)", call, f.receiver, f.DomainName())
	}

	if f.proto.IsTimestamp() {
		call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallTimeToProto)
		return fmt.Sprintf("%s(%s.%s)", call, f.receiver, f.DomainName())
	}

	if f.proto.IsProtoStruct() {
		call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallMapToStruct)
		return fmt.Sprintf("%s(%s.%s)", call, f.receiver, f.DomainName())
	}

	if f.proto.IsMessage() {
		return fmt.Sprintf("%s.%s.IntoWire()", f.receiver, f.DomainName())
	}

	return fmt.Sprintf("%s.%s", f.receiver, f.DomainName())
}

func (f *Field) enumWireType() string {
	var (
		name   = TrimPackageName(f.goType, f.proto.ModuleName())
		prefix string
	)

	// If the enum is from another package we need to add the module name
	// as its prefix.
	module, n, ok := f.handleOtherModuleField(f.goType)
	if ok {
		prefix = ""
		if module != f.proto.ModuleName() {
			prefix = fmt.Sprintf("%s.", module)
		}

		name = fmt.Sprintf("%s%s", prefix, n)
	}

	return fmt.Sprintf("%[1]s.FromString(%[1]s(0), %s.%s)", name, f.receiver, f.goName)
}

func (f *Field) ConvertDomainTypeToArrayWireType(receiver string) string {
	if f.proto.IsEnum() {
		name := TrimPackageName(f.goType, f.proto.ModuleName())
		if module, n, ok := f.handleOtherModuleField(f.goType); ok {
			prefix := ""
			if module != f.proto.ModuleName() {
				prefix = fmt.Sprintf("%s.", module)
			}

			name = fmt.Sprintf("%s%s", prefix, n)
		}

		return fmt.Sprintf("%s.FromString(0, %s)", name, receiver)
	}

	if f.proto.IsTimestamp() {
		call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallTimeToProto)
		return fmt.Sprintf("%s(%s)", call, receiver)
	}

	if f.proto.IsMessage() {
		return fmt.Sprintf("%s.IntoWire()", receiver)
	}

	return receiver
}

func (f *Field) ConvertDomainTypeToMapWireType(receiver string) string {
	_, value, valueKind := f.getMapKeyValueTypesForWire()

	if valueKind.Kind() == protoreflect.EnumKind {
		return fmt.Sprintf("%[1]s.FromString(%[1]s(0), %s)", value, receiver)
	}

	if valueKind.Kind() == protoreflect.MessageKind {
		if strings.Contains(value, "ts.Timestamp") {
			call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallTimeToProto)
			return fmt.Sprintf("%s(%s)", call, receiver)
		}

		return fmt.Sprintf("%s.IntoWire()", receiver)
	}

	return receiver
}

func (f *Field) ConvertWireOutputToOutbound(receiver string) string {
	if f.outbound != nil && f.outbound.GetBitflag() != nil {
		var (
			valuesVar = f.outbound.GetBitflag().GetValues()
			prefix    = f.outbound.GetBitflag().GetPrefix()
		)

		return fmt.Sprintf("currentEnumValues(%s.%s, %s, \"%s\")", receiver, f.goName, valuesVar, prefix)
	}

	if f.proto.IsEnum() {
		return fmt.Sprintf("%s.%s.ValueWithoutPrefix()", receiver, f.goName)
	}

	if f.proto.IsProtoValue() {
		return fmt.Sprintf("%s.%s.AsInterface()", receiver, f.goName)
	}

	if f.proto.IsTimestamp() {
		call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallProtoToTimePtr)
		return fmt.Sprintf("%s(%s.%s)", call, f.receiver, f.DomainName())
	}

	if f.proto.IsProtoStruct() {
		return fmt.Sprintf("%s.%s.AsMap()", receiver, f.goName)
	}

	if f.proto.IsMessage() {
		return fmt.Sprintf("%s.%s.IntoOutboundOrNil()", receiver, f.goName)
	}

	return fmt.Sprintf("%s.%s", receiver, f.goName)
}

func (f *Field) ConvertWireOutputToMapOutbound(receiver string) string {
	v := f.proto.Schema.Desc.MapValue()

	if v.Kind() == protoreflect.EnumKind {
		return fmt.Sprintf("%s.ValueWithoutPrefix()", receiver)
	}

	if v.Kind() == protoreflect.MessageKind {
		return fmt.Sprintf("%s.IntoOutboundOrNil()", receiver)
	}

	return receiver
}

func (f *Field) ConvertWireOutputToArrayOutbound(receiver string) string {
	if f.proto.IsEnum() {
		return fmt.Sprintf("%s.ValueWithoutPrefix()", receiver)
	}

	if f.proto.IsTimestamp() {
		call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallProtoToTimePtr)
		return fmt.Sprintf("%s(%s)", call, receiver)
	}

	if f.proto.IsMessage() {
		return fmt.Sprintf("%s.IntoOutboundOrNil()", receiver)
	}

	return receiver
}

func (f *Field) ValidationName(receiver string) string {
	var address string
	if f.needsAddressNotation() {
		address = "&"
	}

	return fmt.Sprintf("%s%s.%s", address, receiver, f.goName)
}

func (f *Field) needsAddressNotation() bool {
	if !f.isHTTPService {
		// Non HTTP services always need the address notation
		return true
	}

	return !f.isArray && !f.proto.IsProtoStruct() && !f.proto.IsProtobufWrapper() && !f.proto.IsMessage() && !f.proto.IsOptional()
}

func (f *Field) ValidationCall() string {
	return f.validation.ApiCall()
}
