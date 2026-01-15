package context

import (
	"slices"
	"sort"
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/mapping"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/protobuf/extensions"
	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

const (
	outboundTemplateName = "api:outbound"
)

// Message represents a message to be used inside templates by its context.
type Message struct {
	Name         string
	DomainName   string
	WireName     string
	OutboundName string
	Type         mapping.MessageKind
	Fields       []*Field
	ProtoMessage *protobuf.Message
	Mapping      *mapping.Message

	isHTTPService bool
	extensions    *extensions.MikrosMessageExtensions
}

type loadMessagesOptions struct {
	Settings *settings.Settings
}

func loadMessages(pkg *protobuf.Protobuf, opt loadMessagesOptions) ([]*Message, error) {
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
			converter = mapping.NewMessage(mapping.MessageOptions{
				Settings: opt.Settings,
			})
		)

		for i, f := range m.Fields {
			field, err := loadField(loadFieldOptions{
				IsHTTPService:  isHTTPService,
				ModuleName:     pkg.ModuleName,
				Receiver:       getReceiver(m.Name),
				Field:          f,
				Message:        m,
				Endpoint:       endpoint,
				MessageMapping: converter,
				Settings:       opt.Settings,
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
			Mapping:       converter,
			extensions:    extensions.LoadMessageExtensions(m.Proto),
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

// GetReceiverName returns the receiver name for the message.
func (m *Message) GetReceiverName() string {
	return getReceiver(m.Name)
}

// HasArrayField returns true if the message has at least one array field.
func (m *Message) HasArrayField() bool {
	for _, field := range m.Fields {
		if field.IsArray {
			return true
		}
	}

	return false
}

// HasMapField returns true if the message has at least one map field.
func (m *Message) HasMapField() bool {
	for _, field := range m.Fields {
		if field.IsMap {
			return true
		}
	}

	return false
}

// BindableFields returns the fields that can be bound.
func (m *Message) BindableFields(templateName string) []*Field {
	filter := func(field *Field) bool {
		return field.IsBindable()
	}
	if templateName == outboundTemplateName {
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

// ArrayFields returns the fields that are arrays.
func (m *Message) ArrayFields() []*Field {
	var fields []*Field
	for _, field := range m.Fields {
		if field.IsArray && !field.IsMap {
			fields = append(fields, field)
		}
	}

	return fields
}

// DomainExport returns true if the message should be exported to the domain.
func (m *Message) DomainExport() bool {
	if m.extensions != nil {
		if options := m.extensions.GetDomain(); options != nil {
			return !options.GetDontExport()
		}
	}

	return true
}

// OutboundExport returns true if the message should be exported to the outbound
// template.
func (m *Message) OutboundExport() bool {
	// Response messages from HTTP services always have outbound enabled
	if m.Type == mapping.WireOutput && m.isHTTPService {
		return true
	}
	if m.extensions != nil {
		if options := m.extensions.GetOutbound(); options != nil {
			return options.GetExport()
		}
	}

	return false
}

// MapFields returns the fields that are maps.
func (m *Message) MapFields(templateName string) []*Field {
	var fields []*Field
	for _, field := range m.GetFields(templateName) {
		if field.IsMap {
			fields = append(fields, field)
		}
	}

	return fields
}

// HasCustomAPICodeExtension returns true if the message has custom API code
// defined in it.
func (m *Message) HasCustomAPICodeExtension() bool {
	if m.extensions != nil {
		if options := m.extensions.GetCustomApi(); options != nil {
			return len(options.GetFunction()) > 0 || len(options.GetBlock()) > 0
		}
	}

	return false
}

// CustomFunction represents a custom function defined in a message.
type CustomFunction struct {
	Signature string
	Body      string
}

// CustomFunctions returns the custom functions defined in the message.
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

// CustomBlock represents a custom block defined in a message.
type CustomBlock struct {
	Block string
}

// CustomBlocks returns the custom blocks defined in the message.
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

// GetFields returns the fields of the message.
func (m *Message) GetFields(templateName string) []*Field {
	filter := func(field *Field) bool {
		return true
	}
	if templateName == outboundTemplateName {
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

// HasBitflagField returns true if the message has at least one bitflag field.
func (m *Message) HasBitflagField() bool {
	for _, field := range m.Fields {
		if field.IsOutboundBitflag() {
			return true
		}
	}

	return false
}

// HasValidatableField returns true if the message has at least one field that
// is validatable.
func (m *Message) HasValidatableField() bool {
	for _, field := range m.Fields {
		if field.IsValidatable() {
			return true
		}
	}

	return false
}

// ValidationNeedsCustomRuleOptions returns true if the validation needs custom
// rule options.
func (m *Message) ValidationNeedsCustomRuleOptions() bool {
	for _, field := range m.Fields {
		ext := extensions.LoadFieldExtensions(field.ProtoField.Proto)
		if ext == nil {
			continue
		}

		if validation := ext.GetValidate(); validation != nil {
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

// IsWireInputKind returns true if the message is a wire input message.
func (m *Message) IsWireInputKind() bool {
	return m.Type == mapping.WireInput
}

// ValidatableFields returns the fields that are validatable.
func (m *Message) ValidatableFields() []*Field {
	var fields []*Field
	for _, f := range m.Fields {
		if f.IsValidatable() {
			fields = append(fields, f)
		}
	}

	return fields
}
