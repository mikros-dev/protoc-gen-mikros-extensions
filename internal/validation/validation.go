package validation

import (
	"errors"
	"fmt"
	"strings"

	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

// CallOptions represents the options to build a validation call.
type CallOptions struct {
	IsArray   bool
	IsMessage bool
	ProtoName string
	Receiver  string
	ProtoType string
	Options   *extensions.MikrosFieldExtensions
	Settings  *settings.Settings
	Message   *protobuf.Message
}

// Call represents a validation call.
type Call struct {
	apiCall string
}

// NewCall creates a validation call object to retrieve validation expression
// of fields.
func NewCall(options *CallOptions) (*Call, error) {
	var apiCall string
	if options != nil {
		c, err := buildAPICall(options)
		if err != nil {
			return nil, err
		}
		apiCall = c
	}

	return &Call{
		apiCall: apiCall,
	}, nil
}

func buildAPICall(options *CallOptions) (string, error) {
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
		parts             []string
		validationOptions = options.Options.GetValidate()
		begin             = handleBeginCall(options, requiredCondition)
	)

	if begin != "" {
		parts = append(parts, begin)
	}

	// Handle required
	if validationOptions.GetRequired() || requiredCondition != nil {
		req := "validation.Required"
		if msg := validationOptions.GetErrorMessage(); msg != "" {
			req += fmt.Sprintf(`.Error("%s")`, msg)
		}
		parts = append(parts, req)
	}

	// Handle dive/message nesting
	dive, err := buildDiveCall(options)
	if err != nil {
		return "", err
	}
	if dive != "" {
		parts = append(parts, dive)
	}

	// Handle constraints (length, min, max)
	parts = append(parts, buildConstraints(validationOptions)...)

	// Handle rules and finalize
	call := strings.Join(parts, ", ")
	call, err = handleRule(options, call)
	if err != nil {
		return "", err
	}

	return handleEndCall(options, requiredCondition, call), nil
}

func buildDiveCall(options *CallOptions) (string, error) {
	opts := options.Options.GetValidate()
	if !opts.GetDive() {
		return "", nil
	}

	if !options.IsArray && !options.IsMessage {
		return "", fmt.Errorf(
			"field '%s' must be an array or a another message to have dive rule option enabled",
			options.ProtoName,
		)
	}

	var diveParts []string
	if options.IsArray {
		diveParts = append(diveParts, "validation.Each(")
	}
	if options.IsMessage {
		diveParts = append(diveParts, fmt.Sprintf("validation.By(%vValidator(options...)", options.ProtoType))
	}

	return strings.Join(diveParts, ", "), nil
}

func buildConstraints(opts *extensions.FieldValidateOptions) []string {
	var constraints []string
	if opts.GetMaxLength() > 0 {
		constraints = append(constraints, fmt.Sprintf("validation.Length(1, %d)", opts.GetMaxLength()))
	}
	if opts.GetMin() > 0 {
		constraints = append(constraints, fmt.Sprintf("validation.Min(%d)", opts.GetMin()))
	}
	if opts.GetMax() > 0 {
		constraints = append(constraints, fmt.Sprintf("validation.Max(%d)", opts.GetMax()))
	}

	return constraints
}

func handleBeginCall(options *CallOptions, requiredCondition *requiredCondition) string {
	if requiredCondition == nil {
		return ""
	}

	condition := buildConditionalValidationCall(options, requiredCondition)
	return fmt.Sprintf("validation.When(%s", condition)
}

func buildConditionalValidationCall(options *CallOptions, condition *requiredCondition) string {
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

func handleEndCall(options *CallOptions, requiredCondition *requiredCondition, call string) string {
	validationOptions := options.Options.GetValidate()
	if validationOptions.GetDive() || requiredCondition != nil {
		call += ")"
	}

	return call
}

func needsComma(call string) bool {
	return call != "" && !strings.HasSuffix(call, "(")
}

// APICall returns the API call to be used in the generated code.
func (c *Call) APICall() string {
	return c.apiCall
}
