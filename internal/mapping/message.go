package mapping

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
	if m.Kind(name) == WireInput {
		suffix = ""
	}

	return name + suffix
}

// WireToDomainMapValueType is a wrapper for domain mapping in map contexts.
func (m *Message) WireToDomainMapValueType(name string) string {
	return m.WireToDomain(name)
}

// WireToDomain resolves the Go type name for the Domain layer.
func (m *Message) WireToDomain(name string) string {
	if name == "Timestamp" {
		return "time.Time"
	}

	if strings.HasSuffix(name, m.settings.Suffix.Domain) {
		return name
	}

	// Wire to Domain
	old := m.settings.Suffix.Wire
	if m.Kind(name) == WireInput {
		old = m.settings.Suffix.WireInput
	}

	return m.replaceSuffix(name, old, m.settings.Suffix.Domain)
}

// WireOutputToOutbound converts the message wire to the outbound type.
func (m *Message) WireOutputToOutbound(name string) string {
	if strings.HasSuffix(name, m.settings.Suffix.Outbound) {
		return name
	}

	// WireOutput to Outbound
	old := m.settings.Suffix.Wire
	if m.Kind(name) == WireOutput {
		old = m.settings.Suffix.WireOutput
	}

	return m.replaceSuffix(name, old, m.settings.Suffix.Outbound)
}

// Kind identifies which architectural layer a message name belongs to based
// on suffixes.
func (m *Message) Kind(name string) MessageKind {
	switch {
	case strings.HasSuffix(name, m.settings.Suffix.Wire):
		return Wire
	case strings.HasSuffix(name, m.settings.Suffix.WireInput):
		return WireInput
	case strings.HasSuffix(name, m.settings.Suffix.WireOutput):
		return WireOutput
	default:
		return UnknownMessageKind
	}
}

func (m *Message) replaceSuffix(name, old, new string) string {
	if old == "" {
		return name + new
	}

	return strings.ReplaceAll(name, old, new)
}