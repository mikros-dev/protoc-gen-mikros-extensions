package imports

import (
	"fmt"
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/mikros/extensions"
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
			)

			if fieldExtensions != nil {
				validation = fieldExtensions.GetValidate()
			}
			if validation == nil {
				continue
			}

			if validation.GetRule() == mikros_extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_REGEX {
				imports["regex"] = packages["regex"]
				continue
			}

			call := f.ValidationCall
			if cfg.Validations != nil && cfg.Validations.RulePackageImport != nil {
				if strings.Contains(call, fmt.Sprintf("%s.", cfg.Validations.RulePackageImport.Alias)) {
					imports[cfg.Validations.RulePackageImport.Name] = &Import{
						Alias: cfg.Validations.RulePackageImport.Alias,
						Name:  cfg.Validations.RulePackageImport.Name,
					}
				}
			}
		}
	}

	return toSlice(imports)
}
