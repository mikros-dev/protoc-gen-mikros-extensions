package protobuf

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	descriptor "google.golang.org/protobuf/types/descriptorpb"
)

// Message represents a message loaded from protobuf.
type Message struct {
	Name       string
	Fields     []*Field
	Schema     *protogen.Message           `validate:"-"`
	Proto      *descriptor.DescriptorProto `validate:"-"`
	ModuleName string
}

func (m *Message) String() string {
	fields := make([]string, len(m.Fields))
	for i, f := range m.Fields {
		fields[i] = f.String()
	}

	return fmt.Sprintf(`{name:%v, fields:[%v]}`,
		m.Name,
		strings.Join(fields, ","))
}

type parseMessagesOptions struct {
	ModuleName string
	Files      map[string]*protogen.File
}

func parseMessages(options *parseMessagesOptions) []*Message {
	var messages []*Message

	for _, file := range options.Files {
		messages = append(messages, ParseMessagesFromFile(file, options.ModuleName)...)
	}

	return messages
}

// ParseMessagesFromFile parses all messages from a given file.
func ParseMessagesFromFile(file *protogen.File, moduleName string) []*Message {
	messages := make([]*Message, len(file.Proto.MessageType))
	for i, msg := range file.Proto.MessageType {
		messages[i] = parseMessage(msg, file.Messages[i], moduleName)
	}

	return messages
}

func parseMessage(proto *descriptor.DescriptorProto, schema *protogen.Message, moduleName string) *Message {
	fields := make([]*Field, len(proto.Field))
	for i, f := range proto.Field {
		fields[i] = parseField(f, schema.Fields[i], moduleName)
	}

	return &Message{
		Name:       proto.GetName(),
		Fields:     fields,
		Schema:     schema,
		Proto:      proto,
		ModuleName: moduleName,
	}
}
