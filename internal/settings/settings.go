package settings

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"dario.cat/mergo"
	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
)

type Settings struct {
	Debug        bool                   `toml:"debug"`
	Suffix       *Suffix                `toml:"suffix"`
	Database     *Database              `toml:"database"`
	Http         *Http                  `toml:"http"`
	Templates    *Templates             `toml:"templates"`
	Dependencies map[string]*Dependency `toml:"dependencies"`
	Validations  *Validations           `toml:"validations"`
	Addons       *Addons                `toml:"addons"`
}

type Suffix struct {
	Domain     string `toml:"domain" default:"Domain"`
	Outbound   string `toml:"outbound" default:"Outbound"`
	Wire       string `toml:"wire" default:"Wire"`
	WireInput  string `toml:"wire_input" default:"Request"`
	WireOutput string `toml:"wire_output" default:"Response"`
}

type Database struct {
	Kind string `toml:"kind" validate:"oneof=mongo" default:"mongo"`
}

type Http struct {
	Framework string `toml:"framework" validate:"oneof=fasthttp" default:"fasthttp"`
}

type Templates struct {
	Api      bool   `toml:"api" default:"true"`
	Test     bool   `toml:"test" default:"true"`
	TestPath string `toml:"test_path"`
	ApiPath  string `toml:"api_path"`
}

type Dependency struct {
	Import      string                 `toml:"import"`
	Alias       string                 `toml:"alias"`
	PackageName string                 `toml:"package_name"`
	Calls       map[string]interface{} `toml:"calls"`
}

type Validations struct {
	RulePackageImport *Import                    `toml:"rule_package_import"`
	Rule              map[string]*ValidationRule `toml:"rule"`
	Custom            map[string]*ValidationRule `toml:"custom"`
}

type ValidationRule struct {
	ArgsRequired bool   `toml:"args_required"`
	Name         string `toml:"name"`
}

type Import struct {
	Name  string `toml:"name"`
	Alias string `toml:"alias"`
}

type Addons struct {
	Path string `toml:"path"`
}

func LoadSettings(filename string) (*Settings, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var settings Settings
	if err := toml.Unmarshal(file, &settings); err != nil {
		return nil, err
	}

	defaultSettings, err := loadDefaultSettings()
	if err != nil {
		return nil, err
	}

	if err := mergo.Merge(&settings, defaultSettings); err != nil {
		return nil, err
	}

	return &settings, nil
}

func loadDefaultSettings() (*Settings, error) {
	s := &Settings{}
	if err := defaults.Set(s); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Settings) Validate() error {
	validate := validator.New()
	if err := validate.Struct(s); err != nil {
		return err
	}

	return nil
}

func (s *Settings) GetDependencyCall(dep, function string) string {
	if dependency, ok := s.Dependencies[dep]; ok {
		if f, ok := dependency.Calls[function]; ok {
			return f.(string)
		}
	}

	return ""
}

func (s *Settings) GetDependencyModuleName(dep string) string {
	if dependency, ok := s.Dependencies[dep]; ok {
		if dependency.Alias != "" {
			return dependency.Alias
		}

		return dependency.PackageName
	}

	return ""
}

func (s *Settings) IsSupportedCustomValidationRule(ruleName string) error {
	if s.Validations == nil {
		return errors.New("validations settings not set")
	}

	if s.Validations.Custom == nil {
		return errors.New("validations custom rule settings not set")
	}

	if _, ok := s.Validations.Custom[ruleName]; !ok {
		return fmt.Errorf("validations custom rule '%s' not set", ruleName)
	}

	return nil
}

func (s *Settings) GetValidationRule(rule extensions.FieldValidatorRule) (*ValidationRule, error) {
	if s.Validations != nil && s.Validations.Rule != nil {
		name := strings.ToLower(strings.TrimPrefix(rule.String(), "FIELD_VALIDATOR_RULE_"))
		if r, ok := s.Validations.Rule[name]; ok {
			return r, nil
		}

		return nil, fmt.Errorf("could not find settings for validation rule '%s'", rule)
	}

	return nil, nil
}

func (s *Settings) GetValidationCustomRule(name string) (*ValidationRule, error) {
	if s.Validations != nil && s.Validations.Custom != nil {
		if r, ok := s.Validations.Custom[name]; ok {
			return r, nil
		}

		return nil, fmt.Errorf("could not find settings for custom rule '%s'", name)
	}

	return nil, nil
}
