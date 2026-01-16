package context

import (
	"github.com/go-playground/validator/v10"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mapping"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

// FieldMapping represents all supported mapping operations of a field.
type FieldMapping struct {
	validation *mapping.FieldValidation
	tag        *mapping.FieldTag
	naming     *mapping.FieldNaming
	fieldType  *mapping.FieldType
	conversion *mapping.FieldConversion
}

type fieldMappingOptions struct {
	IsHTTPService bool
	Receiver      string
	ProtoField    *protobuf.Field
	Message       *mapping.Message
	ProtoMessage  *protobuf.Message
	Settings      *settings.Settings
}

func newFieldMapping(options *fieldMappingOptions) (*FieldMapping, error) {
	ctx := &mapping.FieldMappingContextOptions{
		ProtoField:   options.ProtoField,
		ProtoMessage: options.ProtoMessage,
		Settings:     options.Settings,
		Validate:     validator.New(),
	}

	fieldType, err := mapping.NewFieldType(&mapping.FieldTypeOptions{
		Message:                    options.Message,
		FieldMappingContextOptions: ctx,
	})
	if err != nil {
		return nil, err
	}

	fieldNaming, err := mapping.NewFieldNaming(&mapping.FieldNamingOptions{
		FieldMappingContextOptions: ctx,
	})
	if err != nil {
		return nil, err
	}

	v, err := mapping.NewFieldValidation(mapping.FieldValidationOptions{
		IsHTTPService:              options.IsHTTPService,
		Receiver:                   options.Receiver,
		FieldNaming:                fieldNaming,
		FieldType:                  fieldType,
		FieldMappingContextOptions: ctx,
	})
	if err != nil {
		return nil, err
	}

	tag, err := mapping.NewFieldTag(&mapping.FieldTagOptions{
		FieldNaming:                fieldNaming,
		FieldMappingContextOptions: ctx,
	})
	if err != nil {
		return nil, err
	}
	conversion, err := mapping.NewFieldConversion(&mapping.FieldConversionOptions{
		MessageReceiver:            options.Receiver,
		FieldNaming:                fieldNaming,
		FieldType:                  fieldType,
		FieldMappingContextOptions: ctx,
	})
	if err != nil {
		return nil, err
	}

	return &FieldMapping{
		validation: v,
		tag:        tag,
		naming:     fieldNaming,
		fieldType:  fieldType,
		conversion: conversion,
	}, nil
}

// Types returns the field type converter.
func (f *FieldMapping) Types() *mapping.FieldType {
	return f.fieldType
}

// Tags returns the field tag converter.
func (f *FieldMapping) Tags() *mapping.FieldTag {
	return f.tag
}

// Naming returns the field naming converter.
func (f *FieldMapping) Naming() *mapping.FieldNaming {
	return f.naming
}

// Conversion returns the field conversion converter.
func (f *FieldMapping) Conversion() *mapping.FieldConversion {
	return f.conversion
}

// Validation returns the field validation converter.
func (f *FieldMapping) Validation() *mapping.FieldValidation {
	return f.validation
}
