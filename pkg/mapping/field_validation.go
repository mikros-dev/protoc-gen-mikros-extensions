package mapping

import (
	"fmt"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/internal/validation"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

// FieldValidationOptions represents the options used to create a new
// FieldValidation.
type FieldValidationOptions struct {
	IsHTTPService bool
	Receiver      string
	FieldNaming   *FieldNaming
	FieldType     *FieldType
	ProtoField    *protobuf.Field
	ProtoMessage  *protobuf.Message
	Settings      *settings.Settings
}

// FieldValidation represents the validation logic for a field.
type FieldValidation struct {
	isHTTPService bool
	validation    *validation.Call
	naming        *FieldNaming
	proto         *protobuf.Field
}

// NewFieldValidation creates a new FieldValidation instance.
func NewFieldValidation(options FieldValidationOptions) (*FieldValidation, error) {
	fieldExtensions := loadFieldExtensions(options.ProtoField)

	call, err := newValidationCall(options, fieldExtensions)
	if err != nil {
		return nil, err
	}

	return &FieldValidation{
		isHTTPService: options.IsHTTPService,
		validation:    call,
		naming:        options.FieldNaming,
		proto:         options.ProtoField,
	}, nil
}

func newValidationCall(
	options FieldValidationOptions,
	ext *extensions.MikrosFieldExtensions,
) (*validation.Call, error) {
	if options.Settings == nil {
		return nil, nil
	}

	return validation.NewCall(&validation.CallOptions{
		IsArray:   options.ProtoField.IsArray(),
		IsMessage: options.ProtoField.IsMessage(),
		ProtoName: options.ProtoField.Name,
		Receiver:  options.Receiver,
		WireType:  options.FieldType.Wire(false),
		Options:   ext,
		Settings:  options.Settings,
		Message:   options.ProtoMessage,
	})
}

// CallFunctionName constructs and returns the validation call name for the
// field.
func (f *FieldValidation) CallFunctionName(receiver string) string {
	var address string
	if f.needsAddressNotation() {
		address = "&"
	}

	return fmt.Sprintf("%s%s.%s", address, receiver, f.naming.GoName())
}

func (f *FieldValidation) needsAddressNotation() bool {
	if !f.isHTTPService {
		// Non-HTTP services always need the address notation
		return true
	}

	return !f.proto.IsArray() &&
		!f.proto.IsProtoStruct() &&
		!f.proto.IsProtobufWrapper() &&
		!f.proto.IsMessage() &&
		!f.proto.IsOptional()
}

// Call retrieves the validation API call from the field's validation
// if it exists.
func (f *FieldValidation) Call() string {
	if f.validation == nil {
		return ""
	}

	return f.validation.APICall()
}
