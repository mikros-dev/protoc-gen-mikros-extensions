package context

import (
	"slices"
	"sort"
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/converters"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mikros_extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
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
	converter     *converters.Message
	extensions    *mikros_extensions.MikrosMessageExtensions
}

type LoadMessagesOptions struct {
	Settings *settings.Settings
}

func loadMessages(pkg *protobuf.Protobuf, opt LoadMessagesOptions) ([]*Message, error) {
	var (
		messages      = make([]*Message, len(pkg.Messages))
		isHTTPService bool
	)

	if pkg.Service != nil {
		isHTTPService = pkg.Service.IsHTTP()
	}

	for i, m := range pkg.Messages {
		var (
			fields    = make([]*Field, len(m.Fields))
			endpoint  = getEndpointFromMessage(m.Name, pkg)
			converter = converters.NewMessage(converters.MessageOptions{
				Settings: opt.Settings,
			})
		)

		for i, f := range m.Fields {
			field, err := loadField(LoadFieldOptions{
				IsHTTPService:    isHTTPService,
				ModuleName:       pkg.ModuleName,
				Receiver:         getReceiver(m.Name),
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
			ProtoMessage:  m,
			isHTTPService: pkg.Service != nil && pkg.Service.IsHTTP(),
			converter:     converter,
			extensions:    mikros_extensions.LoadMessageExtensions(m.Proto),
		}
	}

	// Sort messages by name so it does not affect generated code every
	// time the plugin is executed.
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Name < messages[j].Name
	})

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
	return getReceiver(m.Name)
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
	if m.extensions != nil {
		if options := m.extensions.GetDomain(); options != nil {
			return !options.GetDontExport()
		}
	}

	return true
}

func (m *Message) OutboundExport() bool {
	// Response messages from HTTP services always have outbound enabled
	if m.Type == converters.WireOutputMessage && m.isHTTPService {
		return true
	}
	if m.extensions != nil {
		if options := m.extensions.GetOutbound(); options != nil {
			return options.GetExport()
		}
	}

	return false
}

func (m *Message) MapFields(templateName string) []*Field {
	var fields []*Field
	for _, field := range m.GetFields(templateName) {
		if field.IsMap {
			fields = append(fields, field)
		}
	}

	return fields
}

func (m *Message) HasCustomApiCodeExtension() bool {
	if m.extensions != nil {
		if options := m.extensions.GetCustomApi(); options != nil {
			return len(options.GetFunction()) > 0 || len(options.GetBlock()) > 0
		}
	}

	return false
}

type CustomFunction struct {
	Signature string
	Body      string
}

func (m *Message) CustomFunctions() []*CustomFunction {
	var customCodes []*CustomFunction

	if m.extensions != nil {
		if options := m.extensions.GetCustomApi(); options != nil {
			for _, c := range options.GetFunction() {
				customCodes = append(customCodes, &CustomFunction{
					Signature: c.GetSignature(),
					Body:      c.GetBody(),
				})
			}
		}
	}

	return customCodes
}

type CustomBlock struct {
	Block string
}

func (m *Message) CustomBlocks() []*CustomBlock {
	var customBlocks []*CustomBlock

	if m.extensions != nil {
		if options := m.extensions.GetCustomApi(); options != nil {
			for _, c := range options.GetBlock() {
				customBlocks = append(customBlocks, &CustomBlock{
					Block: c,
				})
			}
		}
	}

	return customBlocks
}

func (m *Message) GetFields(templateName string) []*Field {
	filter := func(field *Field) bool {
		return true
	}
	if templateName == "api:outbound" {
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
		ext := mikros_extensions.LoadFieldExtensions(field.ProtoField.Proto)
		if ext == nil {
			continue
		}

		if validation := ext.GetValidate(); validation != nil {
			nonCustomRules := []mikros_extensions.FieldValidatorRule{
				mikros_extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_REGEX,
				mikros_extensions.FieldValidatorRule_FIELD_VALIDATOR_RULE_UNSPECIFIED,
			}

			if !slices.Contains(nonCustomRules, validation.GetRule()) {
				return true
			}
		}
	}

	return false
}

func (m *Message) IsWireInputKind() bool {
	return m.Type == converters.WireInputMessage
}

func (m *Message) ValidatableFields() []*Field {
	var fields []*Field
	for _, f := range m.Fields {
		if f.IsValidatable() {
			fields = append(fields, f)
		}
	}

	return fields
}
