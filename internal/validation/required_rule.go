package validation

import (
	"errors"
	"fmt"
	"strings"

	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

type RequiredCondition struct {
	Rules                      []*RequiredFieldRuleOptions
	Negative                   bool
	UsePrefixCondition         bool
	DefaultConditionOperation  string
	NegativeConditionOperation string
	RuleConditionOperation     string
}

type RequiredFieldRuleOptions struct {
	FieldName string
	Value     string
	Type      descriptor.FieldDescriptorProto_Type
	TypeName  string
}

type requiredRuleParseOptions struct {
	Condition        string
	Reverse          string
	MaxFieldsToParse int
}

// loadRequiredCondition parses the required field condition, if any. It
// ensures that only one required option is used at the moment.
func loadRequiredCondition(options *CallOptions) (*RequiredCondition, error) {
	var found = 0

	requiredIf, err := parseRequiredRuleOptions(options, &requiredRuleParseOptions{
		Condition:        options.Options.GetRequiredIf(),
		Reverse:          options.Options.GetRequiredIfNot(),
		MaxFieldsToParse: 2,
	})
	if err != nil {
		return nil, err
	}
	if requiredIf != nil {
		found += 1
	}

	requiredWith, err := parseRequiredRuleOptions(options, &requiredRuleParseOptions{
		Condition:        options.Options.GetRequiredWith(),
		Reverse:          options.Options.GetRequiredWithout(),
		MaxFieldsToParse: 1,
	})
	if err != nil {
		return nil, err
	}
	if requiredWith != nil {
		found += 1
	}

	requiredAll, err := parseRequiredRuleOptions(options, &requiredRuleParseOptions{
		Condition: options.Options.GetRequiredAll(),
	})
	if err != nil {
		return nil, err
	}
	if requiredAll != nil {
		found += 1
	}

	requiredAny, err := parseRequiredRuleOptions(options, &requiredRuleParseOptions{
		Condition: options.Options.GetRequiredAny(),
	})
	if err != nil {
		return nil, err
	}
	if requiredAny != nil {
		found += 1
	}

	if found > 1 {
		return nil, errors.New("cannot have more than one 'required_' option for field")
	}

	// Decides which required option to return.
	if requiredIf != nil {
		requiredIf.UsePrefixCondition = true
		requiredIf.DefaultConditionOperation = "=="
		requiredIf.NegativeConditionOperation = "!="
		return requiredIf, nil
	}

	if requiredWith != nil {
		requiredWith.UsePrefixCondition = false
		requiredWith.DefaultConditionOperation = "!="
		requiredWith.NegativeConditionOperation = "=="
		return requiredWith, nil
	}

	if requiredAll != nil {
		requiredAll.UsePrefixCondition = true
		requiredAll.DefaultConditionOperation = "=="
		requiredAll.NegativeConditionOperation = "!="
		requiredAll.RuleConditionOperation = "&&"
		return requiredAll, nil
	}

	if requiredAny != nil {
		requiredAny.UsePrefixCondition = true
		requiredAny.DefaultConditionOperation = "=="
		requiredAny.NegativeConditionOperation = "!="
		requiredAny.RuleConditionOperation = "||"
		return requiredAny, nil
	}

	return nil, nil
}

// parseRequiredRuleOptions parses field required tag options strings into a
// validator.RequiredCondition structure.
func parseRequiredRuleOptions(options *CallOptions, parseOptions *requiredRuleParseOptions) (*RequiredCondition, error) {
	var (
		negative  bool
		fieldRule string
	)

	if parseOptions.Condition != "" {
		fieldRule = parseOptions.Condition
	}
	if parseOptions.Reverse != "" {
		if fieldRule != "" {
			return nil, fmt.Errorf("cannot have both '%s' and '%s' options together", parseOptions.Condition, parseOptions.Reverse)
		}

		fieldRule = parseOptions.Reverse
		negative = true
	}

	// No required option was found. No need to return an error here.
	if fieldRule == "" {
		return nil, nil
	}

	rule, err := parseFieldRule(fieldRule, options, parseOptions.MaxFieldsToParse)
	if err != nil {
		return nil, err
	}

	return &RequiredCondition{
		Rules:    rule,
		Negative: negative,
	}, nil
}

// parseFieldRule parses a strings with formats "field" or "field value"
// into a validator.RequiredFieldRuleOptions structure. It also finds the
// field type inside the protobuf structures to allow later validation.
func parseFieldRule(fieldRule string, options *CallOptions, maxFields int) ([]*RequiredFieldRuleOptions, error) {
	var (
		rules  []*RequiredFieldRuleOptions
		fields [][]string
		parts  = strings.Split(fieldRule, " ")
	)

	// Validates if we have the expected number of fields from fieldRule
	if maxFields != 0 && len(parts) != maxFields {
		return nil, fmt.Errorf("malformed field rule, the number of fields should be '%d' but found '%d'", maxFields, len(parts))
	}
	if maxFields == 0 && (len(parts)%2 != 0) {
		// No maxFields means that the rule should have an even number of parts
		// after splitting it. For example: Field1 value1 Field2 value2 Field3 value3
		return nil, errors.New("malformed field rule, the number of fields and values is invalid, it should be even")
	}

	// Puts together field names and values into a single slice inside another
	// slice. This way we can iterate over them later in an easy way.
	if maxFields == 1 {
		fields = append(fields, []string{parts[0]})
	}
	if maxFields != 1 {
		for i, j := 0, 2; i < len(parts); i, j = i+2, j+2 {
			fields = append(fields, parts[i:j])
		}
	}

	for _, field := range fields {
		name, fieldType, typeName, err := findFieldProtoSpec(field[0], options)
		if err != nil {
			return nil, err
		}

		// All fields must have a value to be evaluated.
		value := ""
		switch fieldType {
		case descriptor.FieldDescriptorProto_TYPE_DOUBLE, descriptor.FieldDescriptorProto_TYPE_FLOAT,
			descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_UINT64,
			descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_FIXED64,
			descriptor.FieldDescriptorProto_TYPE_FIXED32, descriptor.FieldDescriptorProto_TYPE_UINT32,
			descriptor.FieldDescriptorProto_TYPE_SFIXED32, descriptor.FieldDescriptorProto_TYPE_SFIXED64,
			descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SINT64:
			value = "0"

		case descriptor.FieldDescriptorProto_TYPE_BOOL:
			value = "false"

		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			value = "nil"
		}

		if len(field) > 1 {
			value = field[1]

			if value == "$empty" {
				value = ""
			}
		}

		rules = append(rules, &RequiredFieldRuleOptions{
			FieldName: name,
			Type:      fieldType,
			TypeName:  typeName,
			Value:     value,
		})
	}

	return rules, nil
}

// findFieldProtoSpec uses a field name, which can be in a golang format or in
// the protobuf format, to find its real information, such as name and type.
func findFieldProtoSpec(name string, options *CallOptions) (string, descriptor.FieldDescriptorProto_Type, string, error) {
	var (
		message = options.Message.Proto
		schema  = options.Message.Schema
	)

	for f, field := range message.Field {
		if name == schema.Fields[f].GoName || name == field.GetJsonName() || name == field.GetName() {
			return schema.Fields[f].GoName, field.GetType(), field.GetTypeName(), nil
		}
	}

	return "", 0, "", fmt.Errorf("could not find field with name '%s' for validation rule", name)
}
