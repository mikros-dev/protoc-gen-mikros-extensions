package context

import (
	"slices"
	"strings"

	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/converters"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/protobuf"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/internal/settings"
	"github.com/rsfreitas/protoc-gen-mikros-extensions/mikros/extensions"
)

type Message struct {
	Name         string
	DomainName   string
	WireName     string
	OutboundName string
	Type         converters.MessageKind
	Fields       []*Field
	ProtoMessage *protobuf.Message

	isHTTPService bool
	receiver      string
	converter     *converters.Message
}

type LoadMessagesOptions struct {
	Settings *settings.Settings
}

func loadMessages(pkg *protobuf.Protobuf, opt LoadMessagesOptions) ([]*Message, error) {
	messages := make([]*Message, len(pkg.Messages))
	for i, m := range pkg.Messages {
		var (
			isHTTPService bool
			fields        = make([]*Field, len(m.Fields))
			receiver      = getReceiver(m.Name)
			endpoint      = getEndpointFromMessage(m.Name, pkg)
			converter     = converters.NewMessage(converters.MessageOptions{
				Settings: opt.Settings,
			})
		)

		if pkg.Service != nil {
			isHTTPService = pkg.Service.IsHTTP()
		}

		for i, f := range m.Fields {
			field, err := loadField(LoadFieldOptions{
				IsHTTPService:    isHTTPService,
				ModuleName:       pkg.ModuleName,
				Receiver:         receiver,
				Field:            f,
				Message:          m,
				Endpoint:         endpoint,
				MessageConverter: converter,
				Settings:         opt.Settings,
			})
			if err != nil {
				return nil, err
			}

			fields[i] = field
		}

		messages[i] = &Message{
			Name:          m.Name,
			DomainName:    converter.WireToDomain(m.Name),
			WireName:      converter.WireName(m.Name),
			OutboundName:  converter.WireOutputToOutbound(m.Name),
			Type:          converter.Kind(m.Name),
			Fields:        fields,
			isHTTPService: pkg.Service != nil && pkg.Service.IsHTTP(),
			receiver:      receiver,
			converter:     converter,
			ProtoMessage:  m,
		}
	}

	return messages, nil
}

func getReceiver(name string) string {
	r := name[0:1]
	return strings.ToLower(r)
}

func getEndpointFromMessage(msgName string, pkg *protobuf.Protobuf) *Endpoint {
	if pkg.Service != nil {
		for _, m := range pkg.Service.Methods {
			if m.RequestType.Name == msgName {
				return getEndpoint(m)
			}
		}
	}

	return nil
}

func (m *Message) GetReceiverName() string {
	return m.receiver
}

func (m *Message) HasArrayField() bool {
	for _, field := range m.Fields {
		if field.IsArray {
			return true
		}
	}

	return false
}

func (m *Message) HasMapField() bool {
	for _, field := range m.Fields {
		if field.IsMap {
			return true
		}
	}

	return false
}

func (m *Message) BindableFields(templateName string) []*Field {
	filter := func(field *Field) bool {
		return field.IsBindable()
	}
	if templateName == "outbound" {
		filter = func(field *Field) bool {
			return field.IsBindable() && !field.OutboundHide()
		}
	}

	var fields []*Field
	for _, field := range m.Fields {
		if filter(field) {
			fields = append(fields, field)
		}
	}

	return fields
}

func (m *Message) ArrayFields() []*Field {
	var fields []*Field
	for _, field := range m.Fields {
		if field.IsArray && !field.IsMap {
			fields = append(fields, field)
		}
	}

	return fields
}

func (m *Message) DomainExport() bool {
	if options := extensions.LoadMessageDomainOptions(m.ProtoMessage.Proto); options != nil {
		return !options.GetDontExport()
	}

	return true
}

func (m *Message) OutboundExport() bool {
	// Response messages from HTTP services always have outbound enabled
	if m.Type == converters.WireOutputMessage && m.isHTTPService {
		return true
	}
	if options := extensions.LoadMessageOutboundOptions(m.ProtoMessage.Proto); options != nil {
		return options.GetExport()
	}

	return false
}

func (m *Message) MapFields() []*Field {
	var fields []*Field
	for _, field := range m.Fields {
		if field.IsMap {
			fields = append(fields, field)
		}
	}

	return fields
}

func (m *Message) HasWireCustomCodeExtension() bool {
	if options := extensions.LoadMessageWireExtensionOptions(m.ProtoMessage.Proto); options != nil {
		return len(options.GetCustomCode()) > 0
	}

	return false
}

type CustomCode struct {
	Signature string
	Body      string
}

func (m *Message) WireCustomCode() []*CustomCode {
	var customCodes []*CustomCode

	if options := extensions.LoadMessageWireExtensionOptions(m.ProtoMessage.Proto); options != nil {
		for _, c := range options.GetCustomCode() {
			customCodes = append(customCodes, &CustomCode{
				Signature: c.GetSignature(),
				Body:      c.GetBody(),
			})
		}
	}

	return customCodes
}

func (m *Message) GetFields(templateName string) []*Field {
	filter := func(field *Field) bool {
		return true
	}
	if templateName == "outbound" {
		filter = func(field *Field) bool {
			return !field.OutboundHide()
		}
	}

	var fields []*Field
	for _, field := range m.Fields {
		if filter(field) {
			fields = append(fields, field)
		}
	}

	return fields
}

func (m *Message) HasBitflagField() bool {
	for _, field := range m.Fields {
		if field.IsOutboundBitflag() {
			return true
		}
	}

	return false
}

func (m *Message) HasProtobufValueField() bool {
	for _, field := range m.Fields {
		if field.IsProtobufValue() {
			return true
		}
	}

	return false
}

func (m *Message) HasValidatableField() bool {
	for _, field := range m.Fields {
		if field.IsValidatable() {
			return true
		}
	}

	return false
}

func (m *Message) ValidationNeedsCustomRuleOptions() bool {
	for _, field := range m.Fields {
		if validation := extensions.LoadFieldValidate(field.ProtoField.Proto); validation != nil {
			nonCustomRules := []extensions.FieldValidatorRule{
				extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_REGEX,
				extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_UNSPECIFIED,
			}

			if !slices.Contains(nonCustomRules, validation.GetRule()) {
				return true
			}
		}
	}

	return false
}
