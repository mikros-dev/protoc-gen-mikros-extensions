package protobuf

import (
	"strings"
)

// ProtoName represents a name in the form of protobuf package.name or just
// name. If the name is in the form of package.name, the Package field will
// contain the package name.
type ProtoName struct {
	Name      string
	ProtoName string
	Package   string
}

func protoName(s string) *ProtoName {
	var (
		name = s
		pkg  string
	)

	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		pkg = strings.Join(parts[0:len(parts)-2], ".")
		name = parts[len(parts)-1]
	}

	return &ProtoName{
		Name:      name,
		ProtoName: s,
		Package:   pkg,
	}
}

func (n *ProtoName) String() string {
	return n.Name
}
