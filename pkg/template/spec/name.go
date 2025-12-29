package spec

import (
	"fmt"
)

// Name represents a template name.
type Name string

// NewName returns a new template name.
func NewName(prefix, name string) Name {
	return Name(fmt.Sprintf("%s:%s", prefix, name))
}

func (n Name) String() string {
	return string(n)
}
