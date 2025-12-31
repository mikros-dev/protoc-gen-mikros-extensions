package testing

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mapping"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

// Field represents a field for testing templates.
type Field struct {
	isArray        bool
	goType         string
	proto          *protobuf.Field
	settings       *settings.Settings
	fieldConverter *mapping.Field
}

// NewFieldOptions represents the options used to create a new field for testing
// templates.
type NewFieldOptions struct {
	IsArray        bool
	GoType         string
	ProtoField     *protobuf.Field
	Settings       *settings.Settings
	FieldConverter *mapping.Field
}

// NewField creates a new field for testing templates.
func NewField(options *NewFieldOptions) *Field {
	return &Field{
		isArray:        options.IsArray,
		goType:         options.GoType,
		proto:          options.ProtoField,
		settings:       options.Settings,
		fieldConverter: options.FieldConverter,
	}
}

// BindingValue returns the binding expression of a field for testing templates.
func (f *Field) BindingValue(isPointer bool) string {
	if f.proto.IsTimestamp() {
		if f.isArray {
			return "v.([]*time.Time)"
		}

		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr)
		return fmt.Sprintf("%s(v.(time.Time))", call)
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
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr)
		binding = fmt.Sprintf("%s(%v)", call, binding)
	}

	return binding
}

// ValueInitCall returns the value initialization expression of a field for
// testing templates.
func (f *Field) ValueInitCall(isPointer bool) string {
	if c, ok := f.customValueInitCall(); ok {
		return c
	}

	if f.proto.IsProtoValue() {
		return "nil"
	}

	if f.proto.IsMap() || f.isArray || f.proto.IsMessage() {
		return fmt.Sprintf(
			"zeroValue(res.%s).Interface().(%s)",
			f.proto.GoName,
			f.fieldConverter.DomainTypeForTest(isPointer),
		)
	}

	if f.proto.Type == descriptor.FieldDescriptorProto_TYPE_STRING {
		value := `""`
		if f.proto.IsOptional() {
			call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr)
			value = fmt.Sprintf("%s(%s)", call, value)
		}

		return value
	}

	if f.proto.Type == descriptor.FieldDescriptorProto_TYPE_BOOL {
		value := "false"
		if f.proto.IsOptional() {
			call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr)
			value = fmt.Sprintf("%s(%s)", call, value)
		}

		return value
	}

	if f.proto.IsEnum() {
		return f.getEnumTestCallValue()
	}

	value := "0"
	if f.proto.IsOptional() {
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr)
		value = fmt.Sprintf("%s(%s(%s))", call, f.fieldConverter.DomainTypeForTest(false), value)
	}

	return value
}

func (f *Field) customValueInitCall() (string, bool) {
	options := mikros_extensions.LoadFieldExtensions(f.proto.Proto)
	if options == nil || options.GetTesting() == nil {
		return "", false
	}

	testing := options.GetTesting()
	call, err := f.settings.GetTestingCustomRule(testing.GetCustomRule())
	if err != nil {
		return "", false
	}

	c := call.Name + "("
	if call.ArgsRequired {
		for _, arg := range testing.GetRuleArgs() {
			c += fmt.Sprintf(`"%s",`, arg)
		}
	}
	c += ")"

	return c, true
}

func (f *Field) getEnumTestCallValue() string {
	var (
		name               = f.proto.Schema.Enum.GoIdent.GoName
		minN, maxN, values = retrieveEnumValue(f.proto.Schema)
		module             string
	)

	path := strings.Split(string(f.proto.Schema.Enum.GoIdent.GoImportPath), "/")
	module = fmt.Sprintf("%v.", path[len(path)-1])
	conversionCall := fmt.Sprintf(
		`%[1]v%[2]v.FromString(0, %[1]v%[2]v_name[randomIndex(%d, %d, []int{%s})]).ValueWithoutPrefix()`,
		module,
		name,
		minN,
		maxN,
		values,
	)

	if f.isArray {
		conversionCall = fmt.Sprintf("%v%v{%v}", module, name, conversionCall)
	}

	if f.proto.IsOptional() {
		call := f.settings.GetCommonCall(settings.CommonAPIConverters, settings.CommonCallToPtr)
		conversionCall = fmt.Sprintf("%s(%s)", call, conversionCall)
	}

	return conversionCall
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
