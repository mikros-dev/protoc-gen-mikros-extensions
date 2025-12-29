package protobuf

import (
	"fmt"
	"strings"

	"github.com/fatih/camelcase"
	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

// Enum represents an enum loaded from protobuf.
type Enum struct {
	Name   string
	Prefix string
	Values []*EnumEntry
	Proto  *descriptor.EnumDescriptorProto
}

func (e *Enum) String() string {
	entries := make([]string, len(e.Values))
	for i, e := range e.Values {
		entries[i] = e.String()
	}

	return fmt.Sprintf(`{name:%v, prefix:%v, values:[%v]}`,
		e.Name,
		e.Prefix,
		strings.Join(entries, ","))
}

// EnumEntry represents an enum entry loaded from protobuf.
type EnumEntry struct {
	ProtoName string
	Proto     *descriptor.EnumValueDescriptorProto
}

func (e *EnumEntry) String() string {
	return fmt.Sprintf(`proto_name:%v`, e.ProtoName)
}

type parseEnumsOptions struct {
	Files map[string]*protogen.File
}

func parseEnums(options *parseEnumsOptions) []*Enum {
	var enums []*Enum

	for _, file := range options.Files {
		enums = append(enums, ParseEnumsFromFile(file)...)
	}

	return enums
}

// ParseEnumsFromFile parses all enums from a given file.
func ParseEnumsFromFile(file *protogen.File) []*Enum {
	var enums []*Enum

	// Parse all enums declared inside messages
	for _, msg := range file.Proto.MessageType {
		for _, enum := range msg.EnumType {
			enums = append(enums, parseEnumFromMessage(enum, msg))
		}
	}

	// Parse all global enums
	for _, enum := range file.Proto.EnumType {
		enums = append(enums, parseEnum(enum))
	}

	return enums
}

func parseEnumFromMessage(protoEnum *descriptor.EnumDescriptorProto, msg *descriptor.DescriptorProto) *Enum {
	name := fmt.Sprintf("%s_%s", msg.GetName(), protoEnum.GetName())
	prefix := strings.ToUpper(strings.Join(camelcase.Split(protoEnum.GetName()), "_")) + "_"

	return &Enum{
		Name:   name,
		Prefix: prefix,
		Values: parseEnumValues(protoEnum),
		Proto:  protoEnum,
	}
}

func parseEnum(protoEnum *descriptor.EnumDescriptorProto) *Enum {
	name := protoEnum.GetName()
	prefix := strings.ToUpper(strings.Join(camelcase.Split(name), "_")) + "_"

	return &Enum{
		Name:   name,
		Prefix: prefix,
		Values: parseEnumValues(protoEnum),
		Proto:  protoEnum,
	}
}

func parseEnumValues(protoEnum *descriptor.EnumDescriptorProto) []*EnumEntry {
	var entries []*EnumEntry

	for _, protoEntry := range protoEnum.GetValue() {
		entries = append(entries, &EnumEntry{
			ProtoName: protoEntry.GetName(),
			Proto:     protoEntry,
		})
	}

	return entries
}
