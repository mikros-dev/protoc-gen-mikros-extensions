package types

import (
	"fmt"
)

type Name string

func NewName(kind Kind, name string) Name {
	return Name(fmt.Sprintf("%s:%s", kind.String(), name))
}

func (n Name) String() string {
	return string(n)
}
