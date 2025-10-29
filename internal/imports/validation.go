package imports

import (
	"fmt"
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

func loadValidationTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	if ctx.HasValidatableMessage {
		imports["validation"] = packages["validation"]
	}

	for _, m := range ctx.ValidatableMessages {
		if m.ValidationNeedsCustomRuleOptions {
			imports["errors"] = packages["errors"]
		}

		for _, f := range m.Fields {
			var (
				fieldExtensions = mikros_extensions.LoadFieldExtensions(f.ProtoField.Proto)
				validation      *mikros_extensions.FieldValidateOptions
				call            = f.ValidationCall
			)

			if fieldExtensions != nil {
				validation = fieldExtensions.GetValidate()
			}
			if validation == nil {
				continue
			}

			if ok := addRegexForValidationTemplate(imports, validation); ok {
				continue
			}

			if cfg.Validations != nil && cfg.Validations.RulePackageImport != nil {
				if strings.Contains(call, fmt.Sprintf("%s.", cfg.Validations.RulePackageImport.Alias)) {
					imports[cfg.Validations.RulePackageImport.Name] = &Import{
						Alias: cfg.Validations.RulePackageImport.Alias,
						Name:  cfg.Validations.RulePackageImport.Name,
					}
				}
			}

			// If a conditional validation is being made, we check if values used
			// by it belong from an external module.
			if isConditionalValidation(call) {
				values := filterExternalModulesValues(call)
				for _, value := range values {
					moduleName := getModuleName(value)
					imports[moduleName] = importAnotherModule(moduleName, ctx.ModuleName, ctx.FullPath)
				}
			}
		}
	}

	return toSlice(imports)
}

func addRegexForValidationTemplate(imports map[string]*Import, validation *mikros_extensions.FieldValidateOptions) bool {
	if validation.GetRule() == mikros_extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_REGEX {
		imports["regex"] = packages["regex"]
		return true
	}

	return false
}

func isConditionalValidation(call string) bool {
	return strings.HasPrefix(call, "validation.When(") && strings.HasSuffix(call, ", validation.Required)")
}

// filterExternalModulesValues extracts values from a validation.When call
// that references symbols from external modules (i.e., prefixed with module name).
func filterExternalModulesValues(call string) []string {
	call = strings.TrimPrefix(call, "validation.When(")
	call = strings.TrimSuffix(call, ", validation.Required)")

	// By default, we handle the logical operator, treat the whole expression
	// as a single condition.
	pattern := ""
	switch {
	case strings.Contains(call, " && "):
		pattern = " && "
	case strings.Contains(call, " || "):
		pattern = " || "
	}

	var (
		values     []string
		conditions = []string{call}
	)

	if pattern != "" {
		conditions = strings.Split(call, pattern)
	}

	for _, cond := range conditions {
		parts := strings.Split(cond, "==")
		if len(parts) != 2 {
			continue
		}

		value := strings.TrimSpace(parts[1])
		if strings.Contains(value, ".") {
			values = append(values, value)
		}
	}

	return values
}

func getModuleName(s string) string {
	parts := strings.Split(s, ".")
	return parts[0]
}
