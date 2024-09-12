package context

import (
	descriptor "google.golang.org/protobuf/types/descriptorpb"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
)

type Enum struct {
	IsBitflagKind bool
	Name          string
	Prefix        string
	Entries       []*EnumEntry
}

type EnumEntry struct {
	HasEntryDefinition bool
	ProtoName          string
	Name               string
}

func loadEnums(pkg *protobuf.Protobuf) []*Enum {
	enums := make([]*Enum, len(pkg.Enums))
	for i, e := range pkg.Enums {
		enum := &Enum{
			Name:    e.Name,
			Prefix:  e.Prefix,
			Entries: loadEnumEntries(e.Proto),
		}

		if decodingOptions := extensions.LoadEnumDecodingOptions(e.Proto); decodingOptions != nil {
			enum.IsBitflagKind = decodingOptions.GetBitflag()
		}

		enums[i] = enum
	}

	return enums
}

func loadEnumEntries(enum *descriptor.EnumDescriptorProto) []*EnumEntry {
	var entries []*EnumEntry

	for _, protoEntry := range enum.GetValue() {
		var (
			name string
			defs = extensions.LoadEnumEntry(protoEntry)
		)

		if defs != nil {
			name = defs.GetName()
		}

		entries = append(entries, &EnumEntry{
			ProtoName:          protoEntry.GetName(),
			Name:               name,
			HasEntryDefinition: defs != nil,
		})
	}

	return entries
}

func (e *Enum) HasEntryDefinition() bool {
	for _, entry := range e.Entries {
		if entry.HasEntryDefinition {
			return true
		}
	}

	return false
}