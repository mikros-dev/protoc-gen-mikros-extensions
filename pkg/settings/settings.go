package settings

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"dario.cat/mergo"
	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
)

// Settings represents the settings loaded from the configuration file.
type Settings struct {
	Debug       bool         `toml:"debug"`
	Suffix      *Suffix      `toml:"suffix" default:"{}"`
	Database    *Database    `toml:"database" default:"{}"`
	HTTP        *HTTP        `toml:"http" default:"{}"`
	Templates   *Templates   `toml:"templates" default:"{}"`
	Validations *Validations `toml:"validations"`
	Addons      *Addons      `toml:"addons"`
	Testing     *Testing     `toml:"testing" default:"{}"`
}

// Suffix represents the suffixes used in the generated code.
type Suffix struct {
	Domain     string `toml:"domain" default:"Domain"`
	Outbound   string `toml:"outbound" default:"Outbound"`
	Wire       string `toml:"wire" default:"Wire"`
	WireInput  string `toml:"wire_input" default:"Request"`
	WireOutput string `toml:"wire_output" default:"Response"`
}

// Database represents the database used in the generated code.
type Database struct {
	Kind string `toml:"kind" validate:"oneof=mongo gorm" default:"mongo"`
}

// HTTP represents the HTTP framework used in the generated code.
type HTTP struct {
	Framework string `toml:"framework" validate:"oneof=fasthttp" default:"fasthttp"`
}

// Templates represents the templates used in the generated code.
type Templates struct {
	API      bool    `toml:"api" default:"true"`
	Test     bool    `toml:"test" default:"false"`
	TestPath string  `toml:"test_path" default:"test"`
	APIPath  string  `toml:"api_path" default:"go"`
	Common   *Common `toml:"common" default:"{}"`
	Routes   *Routes `toml:"routes" default:"{}"`
}

// Common represents the common operations for all templates used in the
// generated code.
type Common struct {
	Converters bool                   `toml:"converters" default:"true"`
	API        map[string]*Dependency `toml:"api"`
}

// Dependency represents a dependency used in the common operations for
// all templates.
type Dependency struct {
	Import      *Import                `toml:"import"`
	PackageName string                 `toml:"package_name"`
	Calls       map[string]interface{} `toml:"calls"`
}

// Routes represents the routes used in the generated code.
type Routes struct {
	PrefixServiceName bool `toml:"prefix_service_name_in_endpoints"`
}

// Validations represents the validations used in the generated code.
type Validations struct {
	RulePackageImport *Import                `toml:"rule_package_import"`
	Rule              map[string]*CustomCall `toml:"rule"`
	Custom            map[string]*CustomCall `toml:"custom"`
}

// CustomCall represents a custom validation call.
type CustomCall struct {
	ArgsRequired bool   `toml:"args_required"`
	Name         string `toml:"name"`
}

// Import represents a package imported inside the templates.
type Import struct {
	Name  string `toml:"name"`
	Alias string `toml:"alias"`
}

// ModuleName returns the module name of the import.
func (i *Import) ModuleName() string {
	var (
		parts = strings.Split(i.Name, "/")
	)

	prefix := parts[len(parts)-1]
	if isVersionPattern(parts[len(parts)-1]) {
		prefix = parts[len(parts)-2]
	}
	if i.Alias != "" {
		prefix = i.Alias
	}

	return prefix
}

func isVersionPattern(s string) bool {
	re := regexp.MustCompile(`^v\d+$`)
	return re.MatchString(s)
}

// Testing represents the testing settings used in the generated code.
type Testing struct {
	PackageImport *Import                `toml:"package_import"`
	Custom        map[string]*CustomCall `toml:"custom"`
}

// Addons represents the addons used in the generated code.
type Addons struct {
	Path string `toml:"path"`
}

// LoadSettings loads the settings from the configuration file.
func LoadSettings(filename string) (*Settings, error) {
	var settings Settings

	if filename != "" {
		file, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		if err := toml.Unmarshal(file, &settings); err != nil {
			return nil, err
		}
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

// Validate validates the settings.
func (s *Settings) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

// IsSupportedCustomValidationRule checks if a custom validation rule is
// supported or not.
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

// GetValidationRule retrieves the validation rule settings for the specified
// rule.
func (s *Settings) GetValidationRule(rule mikros_extensions.FieldValidatorRule) (*CustomCall, error) {
	if s.Validations != nil && s.Validations.Rule != nil {
		name := strings.ToLower(strings.TrimPrefix(rule.String(), "FIELD_VALIDATOR_RULE_"))
		if r, ok := s.Validations.Rule[name]; ok {
			return r, nil
		}

		return nil, fmt.Errorf("could not find settings for validation rule '%s'", rule)
	}

	return nil, nil
}

// GetValidationCustomRule retrieves the custom validation rule settings by the
// given name from the configuration.
func (s *Settings) GetValidationCustomRule(name string) (*CustomCall, error) {
	if s.Validations != nil && s.Validations.Custom != nil {
		if r, ok := s.Validations.Custom[name]; ok {
			return r, nil
		}

		return nil, fmt.Errorf("could not find settings for custom rule '%s'", name)
	}

	return nil, nil
}

// GetTestingCustomRule retrieves the custom validation rule settings for testing
// templates by the given name from the configuration.
func (s *Settings) GetTestingCustomRule(name string) (*CustomCall, error) {
	if s.Testing != nil && s.Testing.Custom != nil {
		if r, ok := s.Testing.Custom[name]; ok {
			return r, nil
		}

		return nil, fmt.Errorf("could not find settings for custom rule '%s'", name)
	}

	return nil, nil
}

// CommonAPI represents supported common APIs.
type CommonAPI string

// Supported common APIs.
const (
	CommonAPIConverters CommonAPI = "converters"
)

func (c CommonAPI) String() string {
	return string(c)
}

// CommonCall represents a call from a common API.
type CommonCall struct {
	api       CommonAPI
	call      string
	fieldName string
}

// Supported common APIs.
//
//revive:disable:line-length-limit
var (
	CommonCallToPtr          = CommonCall{CommonAPIConverters, "toPtr", "to_ptr"}
	CommonCallToValue        = CommonCall{CommonAPIConverters, "toValue", "to_value"}
	CommonCallProtoToTimePtr = CommonCall{CommonAPIConverters, "protoTimestampToTimePtr", "proto_timestamp_to_go_time_ptr"}
	CommonCallTimeToProto    = CommonCall{CommonAPIConverters, "timeToProtoTimestamp", "go_time_to_proto_timestamp"}
	CommonCallMapToStruct    = CommonCall{CommonAPIConverters, "mapToGrpcStruct", "go_map_to_proto_struct"}
	CommonCallToProtoValue   = CommonCall{CommonAPIConverters, "convertToProtobufValue", "go_interface_to_proto_value"}
)

//revive:enable:line-length-limit

// GetCommonCall returns the call for the specified common API and call.
func (s *Settings) GetCommonCall(apiName CommonAPI, call CommonCall) string {
	if s.Templates.Common.Converters {
		return call.call
	}

	if api, ok := s.Templates.Common.API[apiName.String()]; ok {
		if c, ok := buildDependencyCall(api, call); ok {
			return c
		}
	}

	return ""
}

func buildDependencyCall(d *Dependency, call CommonCall) (string, bool) {
	var prefix string
	if d.Import != nil {
		prefix = d.Import.ModuleName()
	}

	c, ok := d.Calls[call.fieldName]
	if !ok {
		return "", false
	}

	return fmt.Sprintf("%s.%s", prefix, c.(string)), true
}
