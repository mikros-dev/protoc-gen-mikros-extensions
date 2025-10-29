package addon

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/addon"
)

// Addon represents a dynamic plugin integration enabling extended functionality
// at runtime.
type Addon struct {
	// Symbol holds a reference to the addon implementation, which must adhere
	// to specific interfaces.
	Symbol interface{}
}

// LoadAddons loads addon plugins from the specified path, filters files with
// `.so` extensions, and initializes Addon objects.
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

// Addon retrieves the addon implementation from the Symbol field if it
// implements the addon.Addon interface.
func (a *Addon) Addon() addon.Addon {
	if ad, ok := a.Symbol.(addon.Addon); ok {
		return ad
	}

	// Should not fall here because we always load proper Addons symbols.
	return nil
}

// OutboundExtension retrieves the addon implementation from the Symbol field
// if it implements the addon.OutboundExtension interface.
func (a *Addon) OutboundExtension() addon.OutboundExtension {
	if ad, ok := a.Symbol.(addon.OutboundExtension); ok {
		return ad
	}

	return nil
}
