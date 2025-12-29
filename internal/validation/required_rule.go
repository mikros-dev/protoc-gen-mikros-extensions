package validation

import (
	"errors"
	"fmt"
	"strings"

	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

type requiredCondition struct {
	Rules                      []*requiredFieldRuleOptions
	Negative                   bool
	UsePrefixCondition         bool
	DefaultConditionOperation  string
	NegativeConditionOperation string
	RuleConditionOperation     string
}

type requiredFieldRuleOptions struct {
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

type ruleConfig struct {
	parseOpts *requiredRuleParseOptions
	apply     func(*requiredCondition)
}

// loadRequiredCondition parses the required field condition, if any. It
// ensures that only one required option is used at the moment.
func loadRequiredCondition(options *CallOptions) (*requiredCondition, error) {
	var (
		configs = buildRuleConfigs(options)
		result  *requiredCondition
	)

	for _, cfg := range configs {
		cond, err := parseRequiredRuleOptions(options, cfg.parseOpts)
		if err != nil {
			return nil, err
		}
		if cond != nil {
			if result != nil {
				return nil, errors.New("cannot have more than one 'required_' option for field")
			}

			cfg.apply(cond)
			result = cond
		}
	}

	return result, nil
}

func buildRuleConfigs(options *CallOptions) []ruleConfig {
	v := options.Options.GetValidate()

	return []ruleConfig{
		{
			parseOpts: &requiredRuleParseOptions{
				Condition:        v.GetRequiredIf(),
				Reverse:          v.GetRequiredIfNot(),
				MaxFieldsToParse: 2,
			},
			apply: func(rc *requiredCondition) {
				rc.UsePrefixCondition = true
				rc.DefaultConditionOperation = "=="
				rc.NegativeConditionOperation = "!="
			},
		},
		{
			parseOpts: &requiredRuleParseOptions{
				Condition:        v.GetRequiredWith(),
				Reverse:          v.GetRequiredWithout(),
				MaxFieldsToParse: 1,
			},
			apply: func(rc *requiredCondition) {
				rc.UsePrefixCondition = false
				rc.DefaultConditionOperation = "!="
				rc.NegativeConditionOperation = "=="
			},
		},
		{
			parseOpts: &requiredRuleParseOptions{
				Condition: v.GetRequiredAll(),
			},
			apply: func(rc *requiredCondition) {
				rc.UsePrefixCondition = true
				rc.DefaultConditionOperation = "=="
				rc.NegativeConditionOperation = "!="
				rc.RuleConditionOperation = "&&"
			},
		},
		{
			parseOpts: &requiredRuleParseOptions{
				Condition: v.GetRequiredAny(),
			},
			apply: func(rc *requiredCondition) {
				rc.UsePrefixCondition = true
				rc.DefaultConditionOperation = "=="
				rc.NegativeConditionOperation = "!="
				rc.RuleConditionOperation = "||"
			},
		},
	}
}

// parseRequiredRuleOptions parses field required tag options strings into a
// validator.requiredCondition structure.
func parseRequiredRuleOptions(
	options *CallOptions,
	parseOptions *requiredRuleParseOptions,
) (*requiredCondition, error) {
	var (
		negative  bool
		fieldRule string
	)

	if parseOptions.Condition != "" {
		fieldRule = parseOptions.Condition
	}
	if parseOptions.Reverse != "" {
		if fieldRule != "" {
			return nil, fmt.Errorf(
				"cannot have both '%s' and '%s' options together",
				parseOptions.Condition,
				parseOptions.Reverse,
			)
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

	return &requiredCondition{
		Rules:    rule,
		Negative: negative,
	}, nil
}

// parseFieldRule parses a string with formats "field" or "field value"
// into a validator.requiredFieldRuleOptions structure. It also finds the
// field type inside the protobuf structures to allow later validation.
func parseFieldRule(
	fieldRule string,
	options *CallOptions,
	maxFields int,
) ([]*requiredFieldRuleOptions, error) {
	parts := strings.Split(fieldRule, " ")
	if err := validateRuleParts(parts, maxFields); err != nil {
		return nil, err
	}

	var rules []*requiredFieldRuleOptions
	for i := 0; i < len(parts); {
		spec, err := findFieldProtoSpec(parts[i], options)
		if err != nil {
			return nil, err
		}

		// Determine value: use provided value or default based on type
		value := getDefaultValueForType(spec.Type)
		i++ // Move to the next part (potential value)

		if maxFields != 1 && i < len(parts) {
			value = parts[i]
			if value == "$empty" {
				value = ""
			}
			i++ // Move to the next field name
		}

		spec.Value = value
		rules = append(rules, spec)
	}

	return rules, nil
}

func validateRuleParts(parts []string, maxFields int) error {
	if maxFields != 0 && len(parts) != maxFields {
		return fmt.Errorf(
			"malformed field rule, the number of fields should be '%d' but found '%d'",
			maxFields,
			len(parts),
		)
	}
	if maxFields == 0 && (len(parts)%2 != 0) {
		return errors.New("malformed field rule, the number of fields and values is invalid, it should be even")
	}

	return nil
}

func getDefaultValueForType(fieldType descriptor.FieldDescriptorProto_Type) string {
	switch fieldType {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE, descriptor.FieldDescriptorProto_TYPE_FLOAT,
		descriptor.FieldDescriptorProto_TYPE_INT64, descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_FIXED32, descriptor.FieldDescriptorProto_TYPE_UINT32,
		descriptor.FieldDescriptorProto_TYPE_SFIXED32, descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SINT64:
		return "0"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		return "false"
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return "nil"
	default:
		return ""
	}
}

// findFieldProtoSpec uses a field name, which can be in a golang format or in
// the protobuf format, to find its real information, such as name and type.
func findFieldProtoSpec(
	name string,
	options *CallOptions,
) (*requiredFieldRuleOptions, error) {
	var (
		message = options.Message.Proto
		schema  = options.Message.Schema
	)

	for f, field := range message.Field {
		if name == schema.Fields[f].GoName || name == field.GetJsonName() || name == field.GetName() {
			return &requiredFieldRuleOptions{
				FieldName: schema.Fields[f].GoName,
				Type:      field.GetType(),
				TypeName:  field.GetTypeName(),
			}, nil
		}
	}

	return nil, fmt.Errorf("could not find field with name '%s' for validation rule", name)
}
