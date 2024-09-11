package context

import (
	"fmt"
	"strings"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/settings"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
)

func loadValidationTemplateImports(ctx *Context, cfg *settings.Settings) []*Import {
	imports := make(map[string]*Import)

	if ctx.HasValidatableMessage() {
		imports["validation"] = packages["validation"]
	}

	for _, m := range ctx.ValidatableMessages() {
		if m.ValidationNeedsCustomRuleOptions() {
			imports["errors"] = packages["errors"]
		}

		for _, f := range m.Fields {
			validation := extensions.LoadFieldValidate(f.field.Proto)
			if validation == nil {
				continue
			}

			if validation.GetRule() == extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_REGEX {
				imports["regex"] = packages["regex"]
				continue
			}

			call := f.ValidationCall()
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
