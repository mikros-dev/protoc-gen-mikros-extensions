package testing

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/converters"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/settings"
)

type NewFieldOptions struct {
	IsArray        bool
	GoType         string
	ProtoField     *protobuf.Field
	Settings       *settings.Settings
	FieldConverter *converters.Field
}

type Field struct {
	isArray        bool
	goType         string
	proto          *protobuf.Field
	settings       *settings.Settings
	fieldConverter *converters.Field
}

func NewField(options *NewFieldOptions) *Field {
	return &Field{
		isArray:        options.IsArray,
		goType:         options.GoType,
		proto:          options.ProtoField,
		settings:       options.Settings,
		fieldConverter: options.FieldConverter,
	}
}

func (f *Field) BindingValue(isPointer bool) string {
	if f.proto.IsTimestamp() {
		if f.isArray {
			return "v.([]*time.Time)"
		}

		return f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallToPtr)
	}

	if f.proto.IsProtoStruct() {
		return "v.(map[string]interface{})"
	}

	if f.proto.IsMap() || f.isArray || f.proto.IsMessage() {
		outputType := f.fieldConverter.DomainTypeForTest(isPointer)
		return fmt.Sprintf("v.(%v)", outputType)
	}

	t := f.goType
	if f.proto.IsEnum() {
		t = "string"
	}

	binding := fmt.Sprintf("v.(%v)", t)
	if f.proto.IsOptional() {
		call := f.settings.GetCommonCall(settings.CommonApiConverters, settings.CommonCallToPtr)
		binding = fmt.Sprintf("%s(%v)", call, binding)
	}

	return binding
}

func (f *Field) ValueInitCall(isPointer bool) string {
	if f.proto.IsProtoValue() {
		return "nil"
	}

	if f.proto.IsMap() || f.isArray || f.proto.IsMessage() {
		return fmt.Sprintf("zeroValue(res.%s).Interface().(%s)", f.proto.GoName, f.fieldConverter.DomainTypeForTest(isPointer))
	}

	if f.proto.Type == descriptor.FieldDescriptorProto_TYPE_STRING {
		return `""`
	}

	if f.proto.Type == descriptor.FieldDescriptorProto_TYPE_BOOL {
		return "false"
	}

	if f.proto.IsEnum() {
		return f.getEnumTestCallValue()
	}

	return "0"
}

func (f *Field) getEnumTestCallValue() string {
	var (
		name               = f.proto.Schema.Enum.GoIdent.GoName
		minN, maxN, values = retrieveEnumValue(f.proto.Schema)
		module             string
	)

	path := strings.Split(string(f.proto.Schema.Enum.GoIdent.GoImportPath), "/")
	module = fmt.Sprintf("%v.", path[len(path)-1])
	call := fmt.Sprintf(`%[1]v%[2]v.FromString(0, %[1]v%[2]v_name[randomIndex(%d, %d, []int{%s})]).ValueWithoutPrefix()`, module, name, minN, maxN, values)

	if f.isArray {
		call = fmt.Sprintf("%v%v{%v}", module, name, call)
	}

	return call
}

func retrieveEnumValue(protoField *protogen.Field) (int, int, string) {
	var (
		minN    int
		maxN    int
		numbers = make([]int32, len(protoField.Enum.Values))
	)

	for i, n := range protoField.Enum.Values {
		// Do not add UNSPECIFIED enums
		if number := int32(n.Desc.Number()); number != 0 {
			numbers[i] = int32(n.Desc.Number())
		}
	}

	maxN = len(numbers) - 1
	values := make([]string, len(numbers))
	for i, n := range numbers {
		values[i] = fmt.Sprintf("%v", n)
	}

	return minN, maxN, strings.Join(values, ",")
}
