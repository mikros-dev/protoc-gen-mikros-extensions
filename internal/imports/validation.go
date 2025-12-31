package imports

import (
	"fmt"
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/template/spec"
)

// Validation represents the 'api/validation.tmpl' importer
type Validation struct{}

// Name returns the template name.
func (v *Validation) Name() spec.Name {
	return spec.NewName("api", "validation")
}

// Load returns a slice of imports for the template.
func (v *Validation) Load(ctx *Context, cfg *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	if ctx.HasValidatableMessage {
		imports["validation"] = packages["validation"]
	}

	for _, m := range ctx.ValidatableMessages {
		if m.ValidationNeedsCustomRuleOptions {
			imports["errors"] = packages["errors"]
		}

		for _, f := range m.Fields {
			v.processField(ctx, cfg, f, imports)
		}
	}

	return toSlice(imports)
}

func (v *Validation) processField(ctx *Context, cfg *settings.Settings, f *Field, imports map[string]*Import) {
	var (
		fieldExtensions = extensions.LoadFieldExtensions(f.ProtoField.Proto)
		validation      *extensions.FieldValidateOptions
	)

	if fieldExtensions != nil {
		validation = fieldExtensions.GetValidate()
	}
	if validation == nil {
		return
	}

	if ok := v.addRegexForValidationTemplate(imports, validation); ok {
		return
	}

	if cfg.Validations != nil && cfg.Validations.RulePackageImport != nil {
		if strings.Contains(f.ValidationCall, fmt.Sprintf("%s.", cfg.Validations.RulePackageImport.Alias)) {
			imports[cfg.Validations.RulePackageImport.Name] = &Import{
				Alias: cfg.Validations.RulePackageImport.Alias,
				Name:  cfg.Validations.RulePackageImport.Name,
			}
		}
	}

	// If a conditional validation is being made, we check if values used
	// by it belong from an external module.
	v.addExternalModuleImport(ctx, f, imports)
}

func (v *Validation) addRegexForValidationTemplate(
	imports map[string]*Import,
	validation *extensions.FieldValidateOptions,
) bool {
	if validation.GetRule() == extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_REGEX {
		imports["regex"] = packages["regex"]
		return true
	}

	return false
}

func (v *Validation) addExternalModuleImport(ctx *Context, field *Field, imports map[string]*Import) {
	call := field.ValidationCall

	if v.isConditionalValidation(call) {
		values := v.filterExternalModulesValues(call)
		for _, value := range values {
			moduleName := getModuleName(value)
			imports[moduleName] = importAnotherModule(moduleName, ctx.ModuleName, ctx.FullPath)
		}
	}
}

func (v *Validation) isConditionalValidation(call string) bool {
	return strings.HasPrefix(call, "validation.When(") && strings.HasSuffix(call, ", validation.Required)")
}

// filterExternalModulesValues extracts values from a validation.When call
// that references symbols from external modules (i.e., prefixed with module name).
func (v *Validation) filterExternalModulesValues(call string) []string {
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
