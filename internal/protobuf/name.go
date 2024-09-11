package protobuf

import (
	"strings"
)

type Name struct {
	Name      string
	ProtoName string
	Package   string
}

func newName(s string) *Name {
	var (
		name = s
		pkg  string
	)

	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		pkg = strings.Join(parts[0:len(parts)-2], ".")
		name = parts[len(parts)-1]
	}

	return &Name{
		Name:      name,
		ProtoName: s,
		Package:   pkg,
	}
}

func (n *Name) String() string {
	return n.Name
}
