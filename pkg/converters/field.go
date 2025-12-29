package converters

import (
	"fmt"
	"strings"

	"github.com/stoewer/go-strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/validation"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

type conversionMode int

const (
	wireToDomain conversionMode = iota
	wireToOutbound
)

// Field is the object used to make conversions between field types.
type Field struct {
	isArray           bool
	isHTTPService     bool
	goType            string
	goName            string
	receiver          string
	msg               *Message
	db                *Database
	fieldExtensions   *mikros_extensions.MikrosFieldExtensions
	messageExtensions *mikros_extensions.MikrosMessageExtensions
	proto             *protobuf.Field
	settings          *settings.Settings
	validation        *validation.Call
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
		fieldExtensions = mikros_extensions.LoadFieldExtensions(options.ProtoField.Proto)
		isArray         = options.ProtoField.Proto.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
	)

	field := &Field{
		isArray:       isArray,
		isHTTPService: options.IsHTTPService,
		goName:        options.ProtoField.Schema.GoName,
		goType: ProtoTypeToGoType(
			options.ProtoField.Schema.Desc.Kind(),
			options.ProtoField.Proto.GetTypeName(),
			options.ProtoMessage.ModuleName,
		),
		receiver:          options.Receiver,
		msg:               options.Message,
		fieldExtensions:   fieldExtensions,
		messageExtensions: mikros_extensions.LoadMessageExtensions(options.ProtoMessage.Proto),
		proto:             options.ProtoField,
		settings:          options.Settings,
	}
	if options.Settings != nil {
		call, err := validation.NewCall(&validation.CallOptions{
			IsArray:   isArray,
			IsMessage: field.proto.IsMessage(),
			ProtoName: options.ProtoField.Name,
			Receiver:  options.Receiver,
			ProtoType: field.WireType(false),
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

// DomainTypeForTest returns the current field type for testing templates for
// the domain.
func (f *Field) DomainTypeForTest(isPointer bool) string {
	return f.convertFromWireType(isPointer, true, wireToDomain)
}

// OutboundType returns the current field type for the outbound response.
func (f *Field) OutboundType(isPointer bool) string {
	return f.convertFromWireType(isPointer, false, wireToOutbound)
}

func (f *Field) convertFromWireType(isPointer, testMode bool, mode conversionMode) string {
	if mode == wireToOutbound && f.fieldExtensions.GetOutbound().GetCustomType() != "" {
		return f.fieldExtensions.GetOutbound().GetCustomType()
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

	return optional(baseType, f.isArray, isPointer)
}

func (f *Field) getBuiltInType(isPointer bool) (string, bool) {
	switch {
	case f.proto.IsEnum():
		return optional("string", f.isArray, isPointer), true
	case f.proto.IsProtoStruct():
		return optional("map[string]interface{}", f.isArray, false), true
	case f.proto.IsTimestamp():
		return optional("time.Time", f.isArray, isPointer), true
	case f.proto.IsProtoValue():
		return "interface{}", true
	}

	return "", false
}

func (f *Field) getExternalModuleType(isPointer, testMode bool, mode conversionMode) (string, bool) {
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

	return optional(prefix+suffix, f.isArray, isPointer), true
}

func optional(outType string, isArray, isPointer bool) string {
	if isPointer {
		return array("*"+outType, isArray)
	}

	return array(outType, isArray)
}

func array(outType string, isArray bool) string {
	if isArray {
		return "[]" + outType
	}

	return outType
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

		module, _, ok := f.handleOtherModuleField(string(v.Message().FullName()))
		if ok && (module != f.proto.ModuleName() || testMode) {
			valueType = fmt.Sprintf("%s.%s", module, valueType)
		}

		value = "*" + valueType
	}

	if v.Kind() == protoreflect.EnumKind {
		value = "string"
	}

	return key, value
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

// DomainName returns the domain name associated with the field. It is formatted
// in UpperCamelCase if annotations is available, or the Go name.
func (f *Field) DomainName() string {
	if f.fieldExtensions != nil {
		if domain := f.fieldExtensions.GetDomain(); domain != nil {
			if n := domain.GetName(); n != "" {
				return strcase.UpperCamelCase(n)
			}
		}
	}

	return f.goName
}

// DomainTag generates the struct tag string for the field based on its domain
// and naming conventions. It also adds the database struct tag if available.
func (f *Field) DomainTag() string {
	var (
		domain    *mikros_extensions.FieldDomainOptions
		fieldName = strcase.SnakeCase(f.DomainName())
		jsonTag   = ",omitempty"
	)

	if f.messageExtensions != nil {
		if messageDomain := f.messageExtensions.GetDomain(); messageDomain != nil {
			if messageDomain.GetNamingMode() == mikros_extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
				fieldName = strcase.LowerCamelCase(f.DomainName())
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

// InboundTag generates and returns the struct tag string for the inbound
// structure.
func (f *Field) InboundTag() string {
	return fmt.Sprintf("`json:\"%s\"`", f.InboundName())
}

// InboundName returns the inbound name of the field, defaulting to snake_case
// unless overwritten by specific extensions.
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
	fieldName := strcase.SnakeCase(name)
	if f.messageExtensions != nil {
		if messageInbound := f.messageExtensions.GetInbound(); messageInbound != nil {
			if messageInbound.GetNamingMode() == mikros_extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
				fieldName = inboundOutboundCamelCase(name)
			}
		}
	}

	return fieldName
}

// OutboundTag generates the outbound struct tag for the field based on its
// domain name and outbound configuration options.
func (f *Field) OutboundTag() string {
	var (
		outbound *mikros_extensions.FieldOutboundOptions
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
	fieldName := strcase.SnakeCase(name)
	if f.messageExtensions != nil {
		if messageOutbound := f.messageExtensions.GetOutbound(); messageOutbound != nil {
			if messageOutbound.GetNamingMode() == mikros_extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
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

// OutboundName returns the outbound field name.
func (f *Field) OutboundName() string {
	return f.goName
}

// OutboundJSONTagFieldName generates and returns the outbound JSON tag name
// for the field.
func (f *Field) OutboundJSONTagFieldName() string {
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
	fieldName := strcase.SnakeCase(name)
	if f.messageExtensions != nil {
		if messageOutbound := f.messageExtensions.GetOutbound(); messageOutbound != nil {
			if messageOutbound.GetNamingMode() == mikros_extensions.NamingMode_NAMING_MODE_CAMEL_CASE {
				fieldName = inboundOutboundCamelCase(name)
			}
		}
	}

	return fieldName
}

// ConvertToWireType converts the field to its wire-compatible type based on
// protobuf schema and field settings.
func (f *Field) ConvertToWireType(wireInput bool) string {
	if f.proto.IsEnum() {
		return f.enumWireType()
	}

	if f.proto.IsProtoValue() {
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToProtoValue)
		return fmt.Sprintf("%s(%s.%s)", call, f.receiver, f.DomainName())
	}

	if f.proto.IsTimestamp() {
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallTimeToProto)
		return fmt.Sprintf("%s(%s.%s)", call, f.receiver, f.DomainName())
	}

	if f.proto.IsProtoStruct() {
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallMapToStruct)
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
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToValue)
		arg = fmt.Sprintf("%s(%s)", call, arg)
	}

	conversionCall := fmt.Sprintf("%[1]s.FromString(%[1]s(0), %s)", name, arg)
	if f.proto.IsOptional() {
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr)
		conversionCall = fmt.Sprintf("%s(%s)", call, conversionCall)
	}

	return conversionCall
}

// ConvertDomainTypeToWireType converts the domain type into its wire-protocol
// representation.
func (f *Field) ConvertDomainTypeToWireType() string {
	if f.proto.IsEnum() {
		call := fmt.Sprintf("%s.%s.ValueWithoutPrefix()", f.receiver, f.DomainName())
		if f.proto.IsOptional() {
			call = fmt.Sprintf("toPtr(%s)", call)
		}

		return call
	}

	if f.proto.IsProtoValue() {
		return fmt.Sprintf("toDomainInterface(%s.%s)", f.receiver, f.DomainName())
	}

	if f.proto.IsTimestamp() {
		return fmt.Sprintf("toDomainTime(%s.%s)", f.receiver, f.DomainName())
	}

	if f.proto.IsProtoStruct() {
		return fmt.Sprintf("toDomainMap(%s.%s)", f.receiver, f.DomainName())
	}

	if f.proto.IsMessage() {
		return fmt.Sprintf("%s.%s.IntoDomain()", f.receiver, f.DomainName())
	}

	return fmt.Sprintf("%s.%s", f.receiver, f.DomainName())
}

// ConvertDomainTypeToArrayWireType converts the domain-specific representation to
// its array wire format as a string.
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
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallTimeToProto)
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

// ConvertWireTypeToArrayDomainType converts a wire type into a
// domain-specific representation for array fields.
func (f *Field) ConvertWireTypeToArrayDomainType(receiver string) string {
	if f.proto.IsEnum() {
		return fmt.Sprintf("%s.ValueWithoutPrefix()", receiver)
	}

	if f.proto.IsTimestamp() {
		return fmt.Sprintf("toDomainTime(%s)", receiver)
	}

	if f.proto.IsMessage() {
		return fmt.Sprintf("%s.IntoDomain()", receiver)
	}

	return receiver
}

// ConvertDomainTypeToMapWireType converts a domain type to its corresponding
// map wire type representation.
func (f *Field) ConvertDomainTypeToMapWireType(receiver string, wireInput bool) string {
	_, value, valueKind := f.getMapKeyValueTypesForWire()

	if valueKind.Kind() == protoreflect.EnumKind {
		return fmt.Sprintf("%[1]s.FromString(%[1]s(0), %s)", value, receiver)
	}

	if valueKind.Kind() == protoreflect.MessageKind {
		if strings.Contains(value, "ts.Timestamp") {
			call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallTimeToProto)
			return fmt.Sprintf("%s(%s)", call, receiver)
		}

		if wireInput {
			return fmt.Sprintf("%s.IntoWireInput()", receiver)
		}

		return fmt.Sprintf("%s.IntoWire()", receiver)
	}

	return receiver
}

// ConvertWireTypeToMapDomainType converts a wire type value into the corresponding
// map domain type representation.
func (f *Field) ConvertWireTypeToMapDomainType(receiver string) string {
	_, value, valueKind := f.getMapKeyValueTypesForWire()

	if valueKind.Kind() == protoreflect.EnumKind {
		return fmt.Sprintf("%v.ValueWithoutPrefix()", receiver)
	}

	if valueKind.Kind() == protoreflect.MessageKind {
		if strings.Contains(value, "ts.Timestamp") {
			return fmt.Sprintf("toDomainTime(%s)", receiver)
		}

		return fmt.Sprintf("%s.IntoDomain()", receiver)
	}

	return receiver
}

// ConvertWireOutputToOutbound converts the field's wire format output into the
// outbound.
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
			call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr)
			conversionCall = fmt.Sprintf("%s(%s)", call, conversionCall)
		}

		return conversionCall
	}

	if f.proto.IsProtoValue() {
		return fmt.Sprintf("%s.%s.AsInterface()", receiver, f.goName)
	}

	if f.proto.IsTimestamp() {
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallProtoToTimePtr)
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

// ConvertWireOutputToMapOutbound converts the field wire output to its map
// outbound representation.
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

// ConvertWireOutputToArrayOutbound converts the field wire format output into
// an appropriate outbound array representation.
func (f *Field) ConvertWireOutputToArrayOutbound(receiver string) string {
	if f.proto.IsEnum() {
		return fmt.Sprintf("%s.ValueWithoutPrefix()", receiver)
	}

	if f.proto.IsTimestamp() {
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallProtoToTimePtr)
		return fmt.Sprintf("%s(%s)", call, receiver)
	}

	if f.proto.IsProtoStruct() {
		return fmt.Sprintf("%s.AsMap()", receiver)
	}

	if f.proto.IsMessage() {
		return fmt.Sprintf("%s.IntoOutboundOrNil()", receiver)
	}

	return receiver
}

// ValidationName constructs and returns the validation call name for the
// field.
func (f *Field) ValidationName(receiver string) string {
	var address string
	if f.needsAddressNotation() {
		address = "&"
	}

	return fmt.Sprintf("%s%s.%s", address, receiver, f.goName)
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
