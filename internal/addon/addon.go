package addon

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/addon"
)

type Addon struct {
	addon.Addon
}

func LoadAddons(path string) ([]*Addon, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var addons []*Addon
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".so" {
			a, err := loadAddon(filepath.Join(path, f.Name()))
			if err != nil {
				return nil, err
			}

			println("carregou addon:", a.Name())
			addons = append(addons, a)
		}
	}

	return addons, nil
}

func loadAddon(path string) (*Addon, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	obj, err := p.Lookup("Addon")
	if err != nil {
		return nil, err
	}

	a, ok := obj.(addon.Addon)
	if !ok {
		return nil, fmt.Errorf("could not find a proper Addon object inside addon '%s'", path)
	}

	return &Addon{
		Addon: a,
	}, nil
}
