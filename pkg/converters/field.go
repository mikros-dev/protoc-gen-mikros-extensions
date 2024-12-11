package converters

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/validation"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/mikros/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

type conversionMode int

const (
	wireToDomain conversionMode = iota
	wireToOutbound
)

type Field struct {
	isArray           bool
	isHTTPService     bool
	goType            string
	goName            string
	receiver          string
	msg               *Message
	db                *Database
	fieldExtensions   *extensions.MikrosFieldExtensions
	messageExtensions *extensions.MikrosMessageExtensions
	proto             *protobuf.Field
	settings          *settings.Settings
	validation        *validation.Call
}

type FieldOptions struct {
	IsHTTPService bool
	Receiver      string
	ProtoField    *protobuf.Field
	Message       *Message
	ProtoMessage  *protobuf.Message
	Settings      *settings.Settings
}

func NewField(options FieldOptions) (*Field, error) {
	var (
		fieldExtensions = extensions.LoadFieldExtensions(options.ProtoField.Proto)
		isArray         = options.ProtoField.Proto.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
	)

	field := &Field{
		isArray:           isArray,
		isHTTPService:     options.IsHTTPService,
		goName:            options.ProtoField.Schema.GoName,
		goType:            ProtoTypeToGoType(options.ProtoField.Schema.Desc.Kind(), options.ProtoField.Proto.GetTypeName(), options.ProtoMessage.ModuleName),
		receiver:          options.Receiver,
		msg:               options.Message,
		fieldExtensions:   fieldExtensions,
		messageExtensions: extensions.LoadMessageExtensions(options.ProtoMessage.Proto),
		proto:             options.ProtoField,
		settings:          options.Settings,
	}
	if options.Settings != nil {
		call, err := validation.NewCall(&validation.CallOptions{
			IsArray:   isArray,
			ProtoName: options.ProtoField.Name,
			Receiver:  options.Receiver,
			Options:   fieldExtensions,
			Settings:  options.Settings,
			Message:   options.ProtoMessage,
		})
		if err != nil {
			return nil, err
		}

		field.validation = call
		field.db = databaseFromString(options.Settings.Database.Kind, fieldExtensions)
	}

	return field, nil
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
	if mode == wireToOutbound && f.fieldExtensions != nil && f.fieldExtensions.GetOutbound() != nil {
		if t := f.fieldExtensions.GetOutbound().GetCustomType(); t != "" {
			return t
		}
	}

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
	if f.fieldExtensions != nil {
		if outbound := f.fieldExtensions.GetOutbound(); outbound != nil {
			if outbound.GetBitflag() != nil {
				return "[]string"
			}
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
	if f.fieldExtensions != nil {
		if domain := f.fieldExtensions.GetDomain(); domain != nil {
			if n := domain.GetName(); n != "" {
				return strcase.ToCamel(n)
			}
		}
	}

	return f.goName
}

func (f *Field) DomainTag() string {
	var (
		domain    *extensions.FieldDomainOptions
		fieldName = strcase.ToSnake(f.DomainName())
		jsonTag   = ",omitempty"
	)

	if f.messageExtensions != nil {
		if messageDomain := f.messageExtensions.GetDomain(); messageDomain != nil {
			if messageDomain.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
				fieldName = strcase.ToLowerCamel(f.DomainName())
			}
		}
	}

	if f.fieldExtensions != nil {
		domain = f.fieldExtensions.GetDomain()
	}

	if domain != nil {
		if domain.GetAllowEmpty() {
			jsonTag = ""
		}
	}

	tag := fmt.Sprintf("`json:\"%s%s\" %s", fieldName, jsonTag, f.db.Tag(fieldName))
	if domain != nil {
		for _, st := range domain.GetStructTag() {
			tag += fmt.Sprintf(` %s:"%s"`, st.GetName(), st.GetValue())
		}
	}
	tag += "`"

	return tag
}

func (f *Field) InboundTag() string {
	return fmt.Sprintf("`json:\"%s\"`", f.InboundName())
}

func (f *Field) InboundName() string {
	name := f.DomainName()
	if f.fieldExtensions != nil {
		if inbound := f.fieldExtensions.GetInbound(); inbound != nil {
			if n := inbound.GetName(); n != "" {
				name = n
			}
		}
	}

	// Default is snake_case
	fieldName := strcase.ToSnake(name)
	if f.messageExtensions != nil {
		if messageInbound := f.messageExtensions.GetInbound(); messageInbound != nil {
			if messageInbound.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
				fieldName = inboundOutboundCamelCase(name)
			}
		}
	}

	return fieldName
}

func (f *Field) OutboundTag() string {
	var (
		outbound *extensions.FieldOutboundOptions
		name     = f.DomainName()
		jsonTag  = ",omitempty"
	)

	if f.fieldExtensions != nil {
		outbound = f.fieldExtensions.GetOutbound()
	}

	if outbound != nil {
		if n := outbound.GetName(); n != "" {
			name = n
		}
		if outbound.GetAllowEmpty() {
			jsonTag = ""
		}
	}

	// Default is snake_case
	fieldName := strcase.ToSnake(name)
	if f.messageExtensions != nil {
		if messageOutbound := f.messageExtensions.GetOutbound(); messageOutbound != nil {
			if messageOutbound.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
				fieldName = inboundOutboundCamelCase(name)
			}
		}
	}

	tag := fmt.Sprintf("`json:\"%s%s\"", fieldName, jsonTag)
	if outbound != nil {
		for _, st := range outbound.GetStructTag() {
			tag += fmt.Sprintf(` %s:"%s"`, st.GetName(), st.GetValue())
		}
	}
	tag += "`"

	return tag
}

func (f *Field) OutboundName() string {
	return f.goName
}

func (f *Field) OutboundJsonTagFieldName() string {
	var (
		name = f.DomainName()
	)

	if f.fieldExtensions != nil {
		if outbound := f.fieldExtensions.GetOutbound(); outbound != nil {
			if n := outbound.GetName(); n != "" {
				name = n
			}
		}
	}

	// Default is snake_case
	fieldName := strcase.ToSnake(name)
	if f.messageExtensions != nil {
		if messageOutbound := f.messageExtensions.GetOutbound(); messageOutbound != nil {
			if messageOutbound.GetNamingMode() == extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
				fieldName = inboundOutboundCamelCase(name)
			}
		}
	}

	return fieldName
}

func (f *Field) ConvertToWireType(wireInput bool) string {
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
		if wireInput {
			return fmt.Sprintf("%s.%s.IntoWireInput()", f.receiver, f.DomainName())
		}

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

	arg := fmt.Sprintf("%s.%s", f.receiver, f.goName)
	if f.proto.IsOptional() {
		call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallToValue)
		arg = fmt.Sprintf("%s(%s)", call, arg)
	}

	conversionCall := fmt.Sprintf("%[1]s.FromString(%[1]s(0), %s)", name, arg)
	if f.proto.IsOptional() {
		call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallToPtr)
		conversionCall = fmt.Sprintf("%s(%s)", call, conversionCall)
	}

	return conversionCall
}

func (f *Field) ConvertDomainTypeToArrayWireType(receiver string, wireInput bool) string {
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
		if wireInput {
			return fmt.Sprintf("%s.IntoWireInput()", receiver)
		}

		return fmt.Sprintf("%s.IntoWire()", receiver)
	}

	return receiver
}

func (f *Field) ConvertDomainTypeToMapWireType(receiver string, wireInput bool) string {
	_, value, valueKind := f.getMapKeyValueTypesForWire()

	if valueKind.Kind() == protoreflect.EnumKind {
		return fmt.Sprintf("%[1]s.FromString(%[1]s(0), %s)", value, receiver)
	}

	if valueKind.Kind() == protoreflect.MessageKind {
		if strings.Contains(value, "ts.Timestamp") {
			call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallTimeToProto)
			return fmt.Sprintf("%s(%s)", call, receiver)
		}

		if wireInput {
			return fmt.Sprintf("%s.IntoWireInput()", receiver)
		}

		return fmt.Sprintf("%s.IntoWire()", receiver)
	}

	return receiver
}

func (f *Field) ConvertWireOutputToOutbound(receiver string) string {
	if f.fieldExtensions != nil {
		if outbound := f.fieldExtensions.GetOutbound(); outbound != nil && outbound.GetBitflag() != nil {
			var (
				valuesVar = outbound.GetBitflag().GetValues()
				prefix    = outbound.GetBitflag().GetPrefix()
			)

			return fmt.Sprintf("currentEnumValues(%s.%s, %s_name, \"%s\")", receiver, f.goName, valuesVar, prefix)
		}
	}

	if f.proto.IsEnum() {
		conversionCall := fmt.Sprintf("%s.%s.ValueWithoutPrefix()", receiver, f.goName)

		if f.proto.IsOptional() {
			call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallToPtr)
			conversionCall = fmt.Sprintf("%s(%s)", call, conversionCall)
		}

		return conversionCall
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
	if f.validation == nil {
		return ""
	}

	return f.validation.ApiCall()
}
