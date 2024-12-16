package translation

import (
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"regexp"

	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

func RustEndpoint(endpoint string) string {
	re := regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)}`)
	return re.ReplaceAllString(endpoint, `:$1`)
}

func RustFieldType(fieldType descriptor.FieldDescriptorProto_Type, isOptional, isArray bool, messageType string, field *protobuf.Field) string {
	rustType := rustFieldType(fieldType)
	if fieldType == descriptor.FieldDescriptorProto_TYPE_ENUM {
		rustType = rustEnumFieldType()
	}
	if fieldType == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		rustType = rustMessageFieldType(messageType)
		if field.IsTimestamp() {
			rustType = "Option<prost_wkt_types::Timestamp>"
		}
	}

	if isArray {
		rustType = "Vec<" + rustType + ">"
	}
	if isOptional {
		rustType = "Option<" + rustType + ">"
	}

	return rustType
}

func rustEnumFieldType() string {
	return ""
}

func rustMessageFieldType(messageType string) string {
	return messageType
}

func rustFieldType(fieldType descriptor.FieldDescriptorProto_Type) string {
	switch fieldType {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return "f64"
	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return "f32"
	case descriptor.FieldDescriptorProto_TYPE_INT64:
		return "i64"
	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		return "u64"
	case descriptor.FieldDescriptorProto_TYPE_INT32:
		return "i32"
	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		return "ui32"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "bool"
	}

	// Everything else as string?
	return "String"
}
