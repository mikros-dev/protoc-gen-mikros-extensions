package args

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Args struct {
	SettingsFilename string
	flags            flag.FlagSet
}

func NewArgsFromString(s string) (*Args, error) {
	if s == "" {
		return &Args{}, nil
	}

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

func (a *Args) GetPluginName() string {
	return filepath.Base(os.Args[0])
}
