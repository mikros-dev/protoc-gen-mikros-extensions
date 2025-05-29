package translation

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
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
			rustType = "prost_wkt_types::Timestamp"
		}
		if field.IsProtoStruct() {
			rustType = "prost_wkt_types::Struct"
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
	// TODO
	return ""
}

func rustMessageFieldType(messageType string) string {
	messageType = strings.ReplaceAll(messageType, ".", "::")
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

func RustFormatCode(input string, args []string) (string, error) {
	var (
		arguments = append([]string{"--emit", "stdout", "--edition"}, args...)
		stdout    bytes.Buffer
		stderr    bytes.Buffer
	)

	cmd := exec.Command("rustfmt", arguments...)
	cmd.Stdin = bytes.NewBufferString(input)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", errors.New(stderr.String())
	}
	if stderr.Len() > 0 {
		return "", errors.New(stderr.String())
	}

	return stdout.String(), nil
}

func RustHeaderArgument(fieldType descriptor.FieldDescriptorProto_Type, protoName string) string {
	if fieldType == descriptor.FieldDescriptorProto_TYPE_BOOL {
		return fmt.Sprintf(`mikros::http::header::to_bool(context.clone(), &headers, "%s")?`, protoName)
	}

	return fmt.Sprintf(`mikros::http::header::to_string(context.clone(), &headers, "%s")?`, protoName)
}

func RustWireOutputToOutbound(fieldType descriptor.FieldDescriptorProto_Type, protoName, receiver string) string {
	if fieldType == descriptor.FieldDescriptorProto_TYPE_MESSAGE {
		return fmt.Sprintf("%s.%s.unwrap().into()", receiver, protoName)
	}

	return fmt.Sprintf("%s.%s", receiver, protoName)
}
