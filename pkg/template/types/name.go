package types

import "fmt"

type Name string

func NewName(prefix, name string) Name {
	return Name(fmt.Sprintf("%s:%s", prefix, name))
}

func (n Name) String() string {
	return string(n)
}
