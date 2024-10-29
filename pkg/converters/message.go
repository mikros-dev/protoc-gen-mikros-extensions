package converters

import (
	"strings"

	"github.com/mikros-dev/protoc-gen-mikros-extensions/pkg/settings"
)

type Message struct {
	settings *settings.Settings
}

type MessageOptions struct {
	Settings *settings.Settings
}

func NewMessage(options MessageOptions) *Message {
	return &Message{
		settings: options.Settings,
	}
}

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
