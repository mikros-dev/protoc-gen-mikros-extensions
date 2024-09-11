package imports

import (
	"fmt"
	"strings"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/imports"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/settings"
)

func loadValidationTemplateImports(ctx *Context, cfg *settings.Settings) []*imports.Import {
	ipt := make(map[string]*imports.Import)

	if ctx.HasValidatableMessage {
		ipt["validation"] = packages["validation"]
	}

	for _, m := range ctx.ValidatableMessages {
		if m.ValidationNeedsCustomRuleOptions {
			ipt["errors"] = packages["errors"]
		}

		for _, f := range m.Fields {
			validation := extensions.LoadFieldValidate(f.ProtoField.Proto)
			if validation == nil {
				continue
			}

			if validation.GetRule() == extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_REGEX {
				ipt["regex"] = packages["regex"]
				continue
			}

			call := f.ValidationCall
			if cfg.Validations != nil && cfg.Validations.RulePackageImport != nil {
				if strings.Contains(call, fmt.Sprintf("%s.", cfg.Validations.RulePackageImport.Alias)) {
					ipt[cfg.Validations.RulePackageImport.Name] = &imports.Import{
						Alias: cfg.Validations.RulePackageImport.Alias,
						Name:  cfg.Validations.RulePackageImport.Name,
					}
				}
			}
		}
	}

	return toSlice(ipt)
}
