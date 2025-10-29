package converters

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ProtoTypeToGoType converts a protobuf type to a Go type.
func ProtoTypeToGoType(protobufType protoreflect.Kind, messageType, moduleName string) string {
	switch protobufType {
	case protoreflect.BoolKind:
		return "bool"

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"

	case protoreflect.FloatKind:
		return "float32"

	case protoreflect.DoubleKind:
		return "float64"

	case protoreflect.StringKind:
		return "string"

	case protoreflect.BytesKind:
		return "[]byte"

	case protoreflect.MessageKind:
		return messageType

	case protoreflect.EnumKind:
		return TrimPackageName(messageType, moduleName)
	}

	return ""
}
