package validation

import (
	"errors"
	"fmt"
	"strings"

	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/mikros/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

type CallOptions struct {
	IsArray   bool
	ProtoName string
	Receiver  string
	Options   *extensions.MikrosFieldExtensions
	Settings  *settings.Settings
	Message   *protobuf.Message
}

type Call struct {
	apiCall string
}

func NewCall(options *CallOptions) (*Call, error) {
	var apiCall string
	if options != nil {
		c, err := buildApiCall(options)
		if err != nil {
			return nil, err
		}
		apiCall = c
	}

	return &Call{
		apiCall: apiCall,
	}, nil
}

func buildApiCall(options *CallOptions) (string, error) {
	if options.Options == nil || options.Options.GetValidate() == nil {
		// No validation
		return "", nil
	}

	return buildCall(options)
}

func buildCall(options *CallOptions) (string, error) {
	requiredCondition, err := loadRequiredCondition(options)
	if err != nil {
		return "", err
	}

	var (
		call              = handleBeginCall(options, requiredCondition)
		validationOptions = options.Options.GetValidate()
	)

	// Required is always enabled if a required_ condition is being used.
	if validationOptions.GetRequired() || requiredCondition != nil {
		if call != "" {
			call += ", "
		}

		call += "validation.Required"
		if msg := validationOptions.GetErrorMessage(); msg != "" {
			call += fmt.Sprintf(`.Error("%s")`, msg)
		}
	}

	if validationOptions.GetDive() {
		if !options.IsArray {
			return "", fmt.Errorf("field '%s' is not of array type to have dive rule option enabled", options.ProtoName)
		}

		if call != "" {
			call += ", "
		}

		call += "validation.Each("
	}

	if validationOptions.GetMaxLength() > 0 {
		if needsComma(call) {
			call += ", "
		}

		call += fmt.Sprintf("validation.Length(1, %d)", validationOptions.GetMaxLength())
	}

	if validationOptions.GetMin() > 0 {
		if needsComma(call) {
			call += ", "
		}

		call += fmt.Sprintf("validation.Min(%d)", validationOptions.GetMin())
	}

	if validationOptions.GetMax() > 0 {
		if needsComma(call) {
			call += ", "
		}

		call += fmt.Sprintf("validation.Max(%d)", validationOptions.GetMax())
	}

	c, err := handleRule(options, call)
	if err != nil {
		return "", err
	}
	call = c

	return handleEndCall(options, requiredCondition, call), nil
}

func handleBeginCall(options *CallOptions, requiredCondition *RequiredCondition) string {
	if requiredCondition == nil {
		return ""
	}

	condition := buildConditionalValidationCall(options, requiredCondition)
	return fmt.Sprintf("validation.When(%s", condition)
}

func buildConditionalValidationCall(options *CallOptions, condition *RequiredCondition) string {
	var (
		args       = ""
		totalRules = len(condition.Rules)
	)

	operation := condition.DefaultConditionOperation
	if condition.Negative {
		operation = condition.NegativeConditionOperation
	}

	for i, rule := range condition.Rules {
		value := rule.Value
		if rule.Type == descriptor.FieldDescriptorProto_TYPE_STRING {
			value = fmt.Sprintf(`"%s"`, rule.Value)
		}

		args += fmt.Sprintf("%s.%s %s %s", options.Receiver, rule.FieldName,
			operation, value)

		if totalRules > 1 && i != (totalRules-1) {
			args += fmt.Sprintf(" %s ", condition.RuleConditionOperation)
		}
	}

	return args
}

func handleRule(options *CallOptions, call string) (string, error) {
	var (
		validationOptions = options.Options.GetValidate()
		rule              = validationOptions.GetRule()
	)

	if rule == extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_UNSPECIFIED {
		return call, nil
	}

	if needsComma(call) {
		call += ", "
	}

	if rule == extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_REGEX {
		args := validationOptions.GetRuleArgs()
		if len(args) == 0 {
			return "", errors.New("no arguments specified for regex rule")
		}

		call += fmt.Sprintf(`validation.Match(regexp.MustCompile("%s"))`, args[0])
		return call, nil
	}

	ruleSettings, err := options.Settings.GetValidationRule(rule)
	if rule == extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_CUSTOM {
		ruleSettings, err = options.Settings.GetValidationCustomRule(validationOptions.GetCustomRule())
	}
	if err != nil {
		return "", err
	}

	call += fmt.Sprintf("%s(opt.CustomRuleOptions", ruleSettings.Name)
	if ruleSettings.ArgsRequired {
		if len(validationOptions.GetRuleArgs()) == 0 {
			return "", fmt.Errorf("no arguments specified for validation rule '%s'", rule)
		}

		for _, arg := range validationOptions.GetRuleArgs() {
			call += fmt.Sprintf(`, "%s"`, arg)
		}
	}
	call += ")"

	return call, nil
}

func handleEndCall(options *CallOptions, requiredCondition *RequiredCondition, call string) string {
	validationOptions := options.Options.GetValidate()
	if validationOptions.GetDive() || requiredCondition != nil {
		call += ")"
	}

	return call
}

func needsComma(call string) bool {
	return call != "" && !strings.HasSuffix(call, "(")
}

func (c *Call) ApiCall() string {
	return c.apiCall
}
