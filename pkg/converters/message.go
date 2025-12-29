package converters

import (
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

// Message is the object used to make conversions between message types.
type Message struct {
	settings *settings.Settings
}

// MessageOptions represents configuration options for initializing message
// name conversion.
type MessageOptions struct {
	Settings *settings.Settings
}

// NewMessage creates a new message converter.
func NewMessage(options MessageOptions) *Message {
	return &Message{
		settings: options.Settings,
	}
}

// WireName returns the wire name for the message.
func (m *Message) WireName(name string) string {
	if strings.HasSuffix(name, m.settings.Suffix.Wire) {
		return name
	}

	suffix := m.settings.Suffix.Wire
	if m.Kind(name) == WireInputMessage {
		suffix = ""
	}

	return name + suffix
}

// WireToDomainMapValueType gets the message domain map value type.
func (m *Message) WireToDomainMapValueType(name string) string {
	if name == "Timestamp" {
		return "time.Time"
	}

	if strings.HasSuffix(name, m.settings.Suffix.Domain) {
		return name
	}

	// Wire to Domain
	old := m.settings.Suffix.Wire
	if m.Kind(name) == WireInputMessage {
		old = m.settings.Suffix.WireInput
	}

	return strings.ReplaceAll(name, old, m.settings.Suffix.Domain)
}

// WireToDomain gets the message domain type.
func (m *Message) WireToDomain(name string) string {
	if strings.HasSuffix(name, m.settings.Suffix.Domain) {
		return name
	}

	// Wire to Domain
	old := m.settings.Suffix.Wire
	if m.Kind(name) == WireInputMessage {
		old = m.settings.Suffix.WireInput
	}

	return strings.ReplaceAll(name, old, m.settings.Suffix.Domain)
}

// WireOutputToOutbound converts the message wire to the outbound type.
func (m *Message) WireOutputToOutbound(name string) string {
	if strings.HasSuffix(name, m.settings.Suffix.Outbound) {
		return name
	}

	// WireOutput to Outbound
	old := m.settings.Suffix.Wire
	if m.Kind(name) == WireOutputMessage {
		old = m.settings.Suffix.WireOutput
	}

	return strings.ReplaceAll(name, old, m.settings.Suffix.Outbound)
}

// Kind returns the message kind.
func (m *Message) Kind(name string) MessageKind {
	if strings.HasSuffix(name, m.settings.Suffix.Wire) {
		return WireMessage
	}
	if strings.HasSuffix(name, m.settings.Suffix.WireInput) {
		return WireInputMessage
	}
	if strings.HasSuffix(name, m.settings.Suffix.WireOutput) {
		return WireOutputMessage
	}

	return UnknownMessageKind
}
