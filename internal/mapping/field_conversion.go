package mapping

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

// FieldConversionOptions represents the options used to create a new
// FieldConversion.
type FieldConversionOptions struct {
	MessageReceiver string
	GoName          string
	GoType          string
	Protobuf        *protobuf.Field
	Settings        *settings.Settings
	FieldExtensions *extensions.MikrosFieldExtensions
	FieldNaming     *FieldNaming
}

// FieldConversion represents the conversion logic for a field.
type FieldConversion struct {
	messageReceiver string
	goName          string
	goType          string
	proto           *protobuf.Field
	settings        *settings.Settings
	extensions      *extensions.MikrosFieldExtensions
	naming          *FieldNaming
}

func newFieldConversion(options *FieldConversionOptions) *FieldConversion {
	return &FieldConversion{
		messageReceiver: options.MessageReceiver,
		goName:          options.GoName,
		goType:          options.GoType,
		proto:           options.Protobuf,
		settings:        options.Settings,
		extensions:      options.FieldExtensions,
		naming:          options.FieldNaming,
	}
}

// ToWireType converts a field's value to its corresponding wire type representation
// based on its protobuf type and settings.
func (f *FieldConversion) ToWireType(wireInput bool) string {
	if f.proto.IsEnum() {
		return f.enumWireType()
	}

	if f.proto.IsProtoValue() {
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToProtoValue)
		return fmt.Sprintf("%s(%s.%s)", call, f.messageReceiver, f.naming.DomainName())
	}

	if f.proto.IsTimestamp() {
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallTimeToProto)
		return fmt.Sprintf("%s(%s.%s)", call, f.messageReceiver, f.naming.DomainName())
	}

	if f.proto.IsProtoStruct() {
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallMapToStruct)
		return fmt.Sprintf("%s(%s.%s)", call, f.messageReceiver, f.naming.DomainName())
	}

	if f.proto.IsMessage() {
		if wireInput {
			return fmt.Sprintf("%s.%s.IntoWireInput()", f.messageReceiver, f.naming.DomainName())
		}

		return fmt.Sprintf("%s.%s.IntoWire()", f.messageReceiver, f.naming.DomainName())
	}

	return fmt.Sprintf("%s.%s", f.messageReceiver, f.naming.DomainName())
}

func (f *FieldConversion) enumWireType() string {
	var (
		name   = TrimPackageName(f.goType, f.proto.ModuleName())
		prefix string
	)

	// If the enum is from another package, we need to add the module name
	// as its prefix.
	module, n, ok := handleOtherModuleField(f.goType, f.proto)
	if ok {
		prefix = ""
		if module != f.proto.ModuleName() {
			prefix = fmt.Sprintf("%s.", module)
		}

		name = fmt.Sprintf("%s%s", prefix, n)
	}

	arg := fmt.Sprintf("%s.%s", f.messageReceiver, f.goName)
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

// DomainTypeToWireType converts a domain-specific type into its corresponding
// wire format type representation.
func (f *FieldConversion) DomainTypeToWireType() string {
	if f.proto.IsEnum() {
		call := fmt.Sprintf("%s.%s.ValueWithoutPrefix()", f.messageReceiver, f.naming.DomainName())
		if f.proto.IsOptional() {
			call = fmt.Sprintf("toPtr(%s)", call)
		}

		return call
	}

	if f.proto.IsProtoValue() {
		return fmt.Sprintf("toDomainInterface(%s.%s)", f.messageReceiver, f.naming.DomainName())
	}

	if f.proto.IsTimestamp() {
		return fmt.Sprintf("toDomainTime(%s.%s)", f.messageReceiver, f.naming.DomainName())
	}

	if f.proto.IsProtoStruct() {
		return fmt.Sprintf("toDomainMap(%s.%s)", f.messageReceiver, f.naming.DomainName())
	}

	if f.proto.IsMessage() {
		return fmt.Sprintf("%s.%s.IntoDomain()", f.messageReceiver, f.naming.DomainName())
	}

	return fmt.Sprintf("%s.%s", f.messageReceiver, f.naming.DomainName())
}

// DomainTypeToArrayWireType converts a domain type to its array wire type
// representation.
func (f *FieldConversion) DomainTypeToArrayWireType(receiver string, wireInput bool) string {
	if f.proto.IsEnum() {
		name := TrimPackageName(f.goType, f.proto.ModuleName())
		if module, n, ok := handleOtherModuleField(f.goType, f.proto); ok {
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

// WireTypeToArrayDomainType converts a wire type to its corresponding array
// domain type based on the proto field type.
func (f *FieldConversion) WireTypeToArrayDomainType(receiver string) string {
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

// DomainTypeToMapWireType converts a domain type to its corresponding map wire
// type representation.
func (f *FieldConversion) DomainTypeToMapWireType(receiver string, wireInput bool) string {
	_, value, valueKind := getMapKeyValueTypesForWire(f.proto)

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

// WireTypeToMapDomainType converts a wire type representation to its corresponding
// domain type representation.
func (f *FieldConversion) WireTypeToMapDomainType(receiver string) string {
	_, value, valueKind := getMapKeyValueTypesForWire(f.proto)

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

// WireOutputToOutbound converts the field's value from its internal representation
// to an outbound-friendly format.
func (f *FieldConversion) WireOutputToOutbound(receiver string) string {
	if f.extensions != nil {
		if outbound := f.extensions.GetOutbound(); outbound != nil && outbound.GetBitflag() != nil {
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
		return fmt.Sprintf("%s(%s.%s)", call, f.messageReceiver, f.naming.DomainName())
	}

	if f.proto.IsProtoStruct() {
		return fmt.Sprintf("%s.%s.AsMap()", receiver, f.goName)
	}

	if f.proto.IsMessage() {
		return fmt.Sprintf("%s.%s.IntoOutboundOrNil()", receiver, f.goName)
	}

	return fmt.Sprintf("%s.%s", receiver, f.goName)
}

func (f *FieldConversion) WireOutputToMapOutbound(receiver string) string {
	v := f.proto.Schema.Desc.MapValue()

	if v.Kind() == protoreflect.EnumKind {
		return fmt.Sprintf("%s.ValueWithoutPrefix()", receiver)
	}

	if v.Kind() == protoreflect.MessageKind {
		return fmt.Sprintf("%s.IntoOutboundOrNil()", receiver)
	}

	return receiver
}

func (f *FieldConversion) WireOutputToArrayOutbound(receiver string) string {
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
