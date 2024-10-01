package addon

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/pkg/addon"
)

type Addon struct {
	Symbol interface{}
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

	if _, ok := obj.(addon.Addon); !ok {
		return nil, fmt.Errorf("could not find a proper Addon object inside addon '%s'", path)
	}

	return &Addon{
		Symbol: obj,
	}, nil
}

func (a *Addon) Addon() addon.Addon {
	if ad, ok := a.Symbol.(addon.Addon); ok {
		return ad
	}

	// Should not fall here because we always load proper Addons symbols.
	return nil
}

func (a *Addon) OutboundExtension() addon.OutboundExtension {
	if ad, ok := a.Symbol.(addon.OutboundExtension); ok {
		return ad
	}

	return nil
}
