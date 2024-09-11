package args

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Args struct {
	SettingsFilename string `validate:"required"`
	flags            flag.FlagSet
}

func NewArgsFromString(s string) (*Args, error) {
	var (
		args       = &Args{}
		parameters = strings.Split(s, ",")
	)

	for _, param := range parameters {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid plugin argument '%v'", param)
		}

		var (
			key   = parts[0]
			value = parts[1]
		)

		if key == "settings" {
			args.SettingsFilename = value
		}
	}

	return args, nil
}

func NewArgs() *Args {
	o := &Args{}

	o.flags.StringVar(&o.SettingsFilename, "settings", "", "Indicates the settings.toml file to be used.")

	return o
}

func (a *Args) FlagsSet() func(string, string) error {
	return a.flags.Set
}

func (a *Args) Validate() error {
	validate := validator.New()
	if err := validate.Struct(a); err != nil {
		return err
	}

	return nil
}

func (a *Args) GetPluginName() string {
	return filepath.Base(os.Args[0])
}
